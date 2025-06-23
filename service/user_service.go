package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"html/template"
	"math/big"
	"os"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/PhanPhuc2609/be-sign-file/constants"
	"github.com/PhanPhuc2609/be-sign-file/dto"
	"github.com/PhanPhuc2609/be-sign-file/entity"
	"github.com/PhanPhuc2609/be-sign-file/helpers"
	"github.com/PhanPhuc2609/be-sign-file/repository"
	"github.com/PhanPhuc2609/be-sign-file/utils"
	"github.com/google/uuid"
)

type (
	UserService interface {
		Register(ctx context.Context, req dto.UserCreateRequest) (dto.UserResponse, error)
		GetAllUserWithPagination(ctx context.Context, req dto.PaginationRequest) (dto.UserPaginationResponse, error)
		GetUserById(ctx context.Context, userId string) (dto.UserResponse, error)
		GetUserByEmail(ctx context.Context, email string) (dto.UserResponse, error)
		SendVerificationEmail(ctx context.Context, req dto.SendVerificationEmailRequest) error
		VerifyEmail(ctx context.Context, req dto.VerifyEmailRequest) (dto.VerifyEmailResponse, error)
		Update(ctx context.Context, req dto.UserUpdateRequest, userId string) (dto.UserUpdateResponse, error)
		Delete(ctx context.Context, userId string) error
		Verify(ctx context.Context, req dto.UserLoginRequest) (dto.TokenResponse, error)
		RefreshToken(ctx context.Context, req dto.RefreshTokenRequest) (dto.TokenResponse, error)
		RevokeRefreshToken(ctx context.Context, userID string) error
		CreateUserCertificate(ctx context.Context, userId, userEmail, userName string) (certPEM, privPEM, pubPEM string, err error)
		IssueUserCertificate(ctx context.Context, userEmail, userName string) (certPEM, privPEM, pubPEM string, err error)
	}

	userService struct {
		userRepo         repository.UserRepository
		refreshTokenRepo repository.RefreshTokenRepository
		jwtService       JWTService
		db               *gorm.DB
	}
)

func NewUserService(
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	jwtService JWTService,
	db *gorm.DB,
) UserService {
	return &userService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtService:       jwtService,
		db:               db,
	}
}

const (
	LOCAL_URL          = "http://localhost:3000"
	VERIFY_EMAIL_ROUTE = "register/verify_email"
)

func SafeRollback(tx *gorm.DB) {
	if r := recover(); r != nil {
		tx.Rollback()
		// TODO: Do you think that we should panic here?
		// panic(r)
	}
}

func (s *userService) Register(ctx context.Context, req dto.UserCreateRequest) (dto.UserResponse, error) {
	var filename string

	_, flag, err := s.userRepo.CheckEmail(ctx, nil, req.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.UserResponse{}, err
	}

	if flag {
		return dto.UserResponse{}, dto.ErrEmailAlreadyExists
	}

	if req.Image != nil {
		imageId := uuid.New()
		ext := utils.GetExtensions(req.Image.Filename)

		filename = fmt.Sprintf("profile/%s.%s", imageId, ext)
		if err := utils.UploadFile(req.Image, filename); err != nil {
			return dto.UserResponse{}, err
		}
	}

	user := entity.User{
		Name:       req.Name,
		TelpNumber: req.TelpNumber,
		ImageUrl:   filename,
		Role:       constants.ENUM_ROLE_USER,
		Email:      req.Email,
		Password:   req.Password,
		IsVerified: false,
	}

	userReg, err := s.userRepo.Register(ctx, nil, user)
	if err != nil {
		return dto.UserResponse{}, dto.ErrCreateUser
	}

	draftEmail, err := makeVerificationEmail(userReg.Email)
	if err != nil {
		return dto.UserResponse{}, err
	}

	err = utils.SendMail(userReg.Email, draftEmail["subject"], draftEmail["body"])
	if err != nil {
		return dto.UserResponse{}, err
	}

	return dto.UserResponse{
		ID:         userReg.ID.String(),
		Name:       userReg.Name,
		TelpNumber: userReg.TelpNumber,
		ImageUrl:   userReg.ImageUrl,
		Role:       userReg.Role,
		Email:      userReg.Email,
		IsVerified: userReg.IsVerified,
	}, nil
}

func makeVerificationEmail(receiverEmail string) (map[string]string, error) {
	expired := time.Now().Add(time.Hour * 24).Format("2006-01-02 15:04:05")
	plainText := receiverEmail + "_" + expired
	token, err := utils.AESEncrypt(plainText)
	if err != nil {
		return nil, err
	}

	verifyLink := LOCAL_URL + "/" + VERIFY_EMAIL_ROUTE + "?token=" + token

	readHtml, err := os.ReadFile("utils/email-template/base_mail.html")
	if err != nil {
		return nil, err
	}

	data := struct {
		Email  string
		Verify string
	}{
		Email:  receiverEmail,
		Verify: verifyLink,
	}

	tmpl, err := template.New("custom").Parse(string(readHtml))
	if err != nil {
		return nil, err
	}

	var strMail bytes.Buffer
	if err := tmpl.Execute(&strMail, data); err != nil {
		return nil, err
	}

	draftEmail := map[string]string{
		"subject": "Cakno - Go Gin Template",
		"body":    strMail.String(),
	}

	return draftEmail, nil
}

func (s *userService) SendVerificationEmail(ctx context.Context, req dto.SendVerificationEmailRequest) error {
	user, err := s.userRepo.GetUserByEmail(ctx, nil, req.Email)
	if err != nil {
		return dto.ErrEmailNotFound
	}

	draftEmail, err := makeVerificationEmail(user.Email)
	if err != nil {
		return err
	}

	err = utils.SendMail(user.Email, draftEmail["subject"], draftEmail["body"])
	if err != nil {
		return err
	}

	return nil
}

func (s *userService) VerifyEmail(ctx context.Context, req dto.VerifyEmailRequest) (dto.VerifyEmailResponse, error) {
	decryptedToken, err := utils.AESDecrypt(req.Token)
	if err != nil {
		return dto.VerifyEmailResponse{}, dto.ErrTokenInvalid
	}

	if !strings.Contains(decryptedToken, "_") {
		return dto.VerifyEmailResponse{}, dto.ErrTokenInvalid
	}

	decryptedTokenSplit := strings.Split(decryptedToken, "_")
	email := decryptedTokenSplit[0]
	expired := decryptedTokenSplit[1]

	now := time.Now()
	expiredTime, err := time.Parse("2006-01-02 15:04:05", expired)
	if err != nil {
		return dto.VerifyEmailResponse{}, dto.ErrTokenInvalid
	}

	if expiredTime.Sub(now) < 0 {
		return dto.VerifyEmailResponse{
			Email:      email,
			IsVerified: false,
		}, dto.ErrTokenExpired
	}

	user, err := s.userRepo.GetUserByEmail(ctx, nil, email)
	if err != nil {
		return dto.VerifyEmailResponse{}, dto.ErrUserNotFound
	}

	if user.IsVerified {
		return dto.VerifyEmailResponse{}, dto.ErrAccountAlreadyVerified
	}

	updatedUser, err := s.userRepo.Update(
		ctx, nil, entity.User{
			ID:         user.ID,
			IsVerified: true,
		},
	)
	if err != nil {
		return dto.VerifyEmailResponse{}, dto.ErrUpdateUser
	}

	return dto.VerifyEmailResponse{
		Email:      email,
		IsVerified: updatedUser.IsVerified,
	}, nil
}

func (s *userService) GetAllUserWithPagination(
	ctx context.Context,
	req dto.PaginationRequest,
) (dto.UserPaginationResponse, error) {
	dataWithPaginate, err := s.userRepo.GetAllUserWithPagination(ctx, nil, req)
	if err != nil {
		return dto.UserPaginationResponse{}, err
	}

	var datas []dto.UserResponse
	for _, user := range dataWithPaginate.Users {
		data := dto.UserResponse{
			ID:         user.ID.String(),
			Name:       user.Name,
			Email:      user.Email,
			Role:       user.Role,
			TelpNumber: user.TelpNumber,
			ImageUrl:   user.ImageUrl,
			IsVerified: user.IsVerified,
		}

		datas = append(datas, data)
	}

	return dto.UserPaginationResponse{
		Data: datas,
		PaginationResponse: dto.PaginationResponse{
			Page:    dataWithPaginate.Page,
			PerPage: dataWithPaginate.PerPage,
			MaxPage: dataWithPaginate.MaxPage,
			Count:   dataWithPaginate.Count,
		},
	}, nil
}

func (s *userService) GetUserById(ctx context.Context, userId string) (dto.UserResponse, error) {
	user, err := s.userRepo.GetUserById(ctx, nil, userId)
	if err != nil {
		return dto.UserResponse{}, dto.ErrGetUserById
	}

	return dto.UserResponse{
		ID:         user.ID.String(),
		Name:       user.Name,
		TelpNumber: user.TelpNumber,
		Role:       user.Role,
		Email:      user.Email,
		ImageUrl:   user.ImageUrl,
		IsVerified: user.IsVerified,
	}, nil
}

func (s *userService) GetUserByEmail(ctx context.Context, email string) (dto.UserResponse, error) {
	emails, err := s.userRepo.GetUserByEmail(ctx, nil, email)
	if err != nil {
		return dto.UserResponse{}, dto.ErrGetUserByEmail
	}

	return dto.UserResponse{
		ID:         emails.ID.String(),
		Name:       emails.Name,
		TelpNumber: emails.TelpNumber,
		Role:       emails.Role,
		Email:      emails.Email,
		ImageUrl:   emails.ImageUrl,
		IsVerified: emails.IsVerified,
	}, nil
}

func (s *userService) Update(ctx context.Context, req dto.UserUpdateRequest, userId string) (
	dto.UserUpdateResponse,
	error,
) {
	user, err := s.userRepo.GetUserById(ctx, nil, userId)
	if err != nil {
		return dto.UserUpdateResponse{}, dto.ErrUserNotFound
	}

	data := entity.User{
		ID:         user.ID,
		Name:       req.Name,
		TelpNumber: req.TelpNumber,
		Role:       user.Role,
		Email:      req.Email,
	}

	userUpdate, err := s.userRepo.Update(ctx, nil, data)
	if err != nil {
		return dto.UserUpdateResponse{}, dto.ErrUpdateUser
	}

	return dto.UserUpdateResponse{
		ID:         userUpdate.ID.String(),
		Name:       userUpdate.Name,
		TelpNumber: userUpdate.TelpNumber,
		Role:       userUpdate.Role,
		Email:      userUpdate.Email,
		IsVerified: user.IsVerified,
	}, nil
}

func (s *userService) Delete(ctx context.Context, userId string) error {
	tx := s.db.Begin()
	defer SafeRollback(tx)

	user, err := s.userRepo.GetUserById(ctx, nil, userId)
	if err != nil {
		return dto.ErrUserNotFound
	}

	err = s.userRepo.Delete(ctx, nil, user.ID.String())
	if err != nil {
		return dto.ErrDeleteUser
	}

	return nil
}

func (s *userService) Verify(ctx context.Context, req dto.UserLoginRequest) (dto.TokenResponse, error) {
	tx := s.db.Begin()
	defer SafeRollback(tx)

	user, err := s.userRepo.GetUserByEmail(ctx, tx, req.Email)
	if err != nil {
		tx.Rollback()
		return dto.TokenResponse{}, errors.New("invalid email or password")
	}

	checkPassword, err := helpers.CheckPassword(user.Password, []byte(req.Password))
	if err != nil || !checkPassword {
		tx.Rollback()
		return dto.TokenResponse{}, errors.New("invalid email or password")
	}

	accessToken := s.jwtService.GenerateAccessToken(user.ID.String(), user.Role)

	refreshTokenString, expiresAt := s.jwtService.GenerateRefreshToken()

	hashedToken, err := helpers.HashPassword(refreshTokenString)
	if err != nil {
		tx.Rollback()
		return dto.TokenResponse{}, err
	}

	if err := s.refreshTokenRepo.DeleteByUserID(ctx, tx, user.ID.String()); err != nil {
		tx.Rollback()
		return dto.TokenResponse{}, err
	}

	refreshToken := entity.RefreshToken{
		UserID:    user.ID,
		Token:     hashedToken,
		ExpiresAt: expiresAt,
	}

	if _, err := s.refreshTokenRepo.Create(ctx, tx, refreshToken); err != nil {
		tx.Rollback()
		return dto.TokenResponse{}, err
	}

	if err := tx.Commit().Error; err != nil {
		return dto.TokenResponse{}, err
	}

	return dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
		Role:         user.Role,
	}, nil
}

func (s *userService) RefreshToken(ctx context.Context, req dto.RefreshTokenRequest) (dto.TokenResponse, error) {
	tx := s.db.Begin()
	defer SafeRollback(tx)

	// Find the refresh token in the database
	dbToken, err := s.refreshTokenRepo.FindByToken(ctx, tx, req.RefreshToken)
	if err != nil {
		tx.Rollback()
		return dto.TokenResponse{}, errors.New(dto.MESSAGE_FAILED_INVALID_REFRESH_TOKEN)
	}

	if time.Now().After(dbToken.ExpiresAt) {
		tx.Rollback()
		return dto.TokenResponse{}, errors.New(dto.MESSAGE_FAILED_EXPIRED_REFRESH_TOKEN)
	}

	user, err := s.userRepo.GetUserById(ctx, tx, dbToken.UserID.String())
	if err != nil {
		tx.Rollback()
		return dto.TokenResponse{}, dto.ErrUserNotFound
	}

	accessToken := s.jwtService.GenerateAccessToken(user.ID.String(), user.Role)

	refreshTokenString, expiresAt := s.jwtService.GenerateRefreshToken()

	hashedToken, err := helpers.HashPassword(refreshTokenString)
	if err != nil {
		tx.Rollback()
		return dto.TokenResponse{}, err
	}

	if err := s.refreshTokenRepo.DeleteByUserID(ctx, tx, user.ID.String()); err != nil {
		tx.Rollback()
		return dto.TokenResponse{}, err
	}

	refreshToken := entity.RefreshToken{
		UserID:    user.ID,
		Token:     hashedToken,
		ExpiresAt: expiresAt,
	}

	if _, err := s.refreshTokenRepo.Create(ctx, tx, refreshToken); err != nil {
		tx.Rollback()
		return dto.TokenResponse{}, err
	}

	if err := tx.Commit().Error; err != nil {
		return dto.TokenResponse{}, err
	}

	return dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
		Role:         user.Role,
	}, nil
}

func (s *userService) RevokeRefreshToken(ctx context.Context, userID string) error {
	tx := s.db.Begin()
	defer SafeRollback(tx)

	// Check if user exists
	_, err := s.userRepo.GetUserById(ctx, tx, userID)
	if err != nil {
		tx.Rollback()
		return dto.ErrUserNotFound
	}

	// Delete all refresh tokens for the user
	if err := s.refreshTokenRepo.DeleteByUserID(ctx, tx, userID); err != nil {
		tx.Rollback()
		return err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

func (s *userService) CreateUserCertificate(ctx context.Context, userId, userEmail, userName string) (certPEM, privPEM, pubPEM string, err error) {
	// 1. Generate RSA key pair
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", "", err
	}
	// 2. Create certificate template
	tmpl := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			CommonName:   userName,
			Organization: []string{"VinCSS User"},
			ExtraNames: []pkix.AttributeTypeAndValue{
				{Type: []int{1, 2, 840, 113549, 1, 9, 1}, Value: userEmail}, // email OID
			},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(1, 0, 0), // 1 year
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	// 3. Self-sign (hoặc dùng CA riêng nếu có)
	certDER, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &priv.PublicKey, priv)
	if err != nil {
		return "", "", "", err
	}
	// 4. Encode to PEM
	certPEM = string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER}))
	privPEM = string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)}))
	pubASN1, _ := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	pubPEM = string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubASN1}))

	// 5. Lưu vào DB
	user, err := s.userRepo.GetUserById(ctx, nil, userId)
	if err != nil {
		return certPEM, privPEM, pubPEM, err
	}
	user.CertPEM = certPEM
	user.PrivPEM = privPEM
	user.PubPEM = pubPEM
	_, err = s.userRepo.Update(ctx, nil, user)
	if err != nil {
		return certPEM, privPEM, pubPEM, err
	}
	return certPEM, privPEM, pubPEM, nil
}

// IssueUserCertificate: CA cấp chứng chỉ cho user, trả về cert, private key, public key (KHÔNG lưu vào DB)
func (s *userService) IssueUserCertificate(ctx context.Context, userEmail, userName string) (certPEM, privPEM, pubPEM string, err error) {
	// 1. Sinh keypair cho user
	userPriv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", "", err
	}
	// 2. Tạo certificate request (CSR) cho user (bỏ qua bước này, tạo trực tiếp cert template)
	tmpl := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			CommonName:   userName,
			Organization: []string{"VinCSS User"},
			ExtraNames: []pkix.AttributeTypeAndValue{
				{Type: []int{1, 2, 840, 113549, 1, 9, 1}, Value: userEmail},
			},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(1, 0, 0),
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	// 3. CA key/cert (demo: sinh tạm CA key/cert, thực tế nên lưu CA key/cert riêng)
	caPriv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", "", err
	}
	caTmpl := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			CommonName: "VinCSS CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
	}
	caCertDER, err := x509.CreateCertificate(rand.Reader, &caTmpl, &caTmpl, &caPriv.PublicKey, caPriv)
	if err != nil {
		return "", "", "", err
	}
	caCert, err := x509.ParseCertificate(caCertDER)
	if err != nil {
		return "", "", "", err
	}
	// 4. Ký cert user bằng CA
	certDER, err := x509.CreateCertificate(rand.Reader, &tmpl, caCert, &userPriv.PublicKey, caPriv)
	if err != nil {
		return "", "", "", err
	}
	// 5. Encode PEM
	certPEM = string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER}))
	privPEM = string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(userPriv)}))
	pubASN1, _ := x509.MarshalPKIXPublicKey(&userPriv.PublicKey)
	pubPEM = string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubASN1}))
	return certPEM, privPEM, pubPEM, nil
}

// IssueCertificateFromCSR: Nhận CSR PEM, CA ký và trả về certificate PEM
func (s *userService) IssueCertificateFromCSR(ctx context.Context, csrPEM string) (certPEM string, err error) {
	// 1. Parse CSR
	block, _ := pem.Decode([]byte(csrPEM))
	if block == nil || block.Type != "CERTIFICATE REQUEST" {
		return "", errors.New("invalid CSR PEM")
	}
	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		return "", err
	}
	if err := csr.CheckSignature(); err != nil {
		return "", errors.New("CSR signature invalid")
	}
	// 2. CA key/cert (demo: sinh tạm CA key/cert, thực tế nên lưu riêng)
	caPriv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", err
	}
	caTmpl := x509.Certificate{
		SerialNumber:          big.NewInt(time.Now().UnixNano()),
		Subject:               pkix.Name{CommonName: "VinCSS CA"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
	}
	caCertDER, err := x509.CreateCertificate(rand.Reader, &caTmpl, &caTmpl, &caPriv.PublicKey, caPriv)
	if err != nil {
		return "", err
	}
	caCert, err := x509.ParseCertificate(caCertDER)
	if err != nil {
		return "", err
	}
	// 3. Tạo cert cho user từ CSR
	tmpl := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      csr.Subject,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	certDER, err := x509.CreateCertificate(rand.Reader, &tmpl, caCert, csr.PublicKey, caPriv)
	if err != nil {
		return "", err
	}
	certPEM = string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER}))
	return certPEM, nil
}

// GenerateCSRWithPublicKey: Nhận public key, tạo CSR PEM cho user
func (s *userService) GenerateCSRWithPublicKey(commonName, email string, pubKey *rsa.PublicKey) (csrPEM string, err error) {
	// 1. Tạo CSR template
	tmpl := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{"VinCSS User"},
		},
		EmailAddresses: []string{email},
	}
	csrDER, err := x509.CreateCertificateRequest(rand.Reader, &tmpl, nil)
	if err != nil {
		return "", err
	}
	// Thay thế public key trong CSR bằng pubKey truyền vào
	csr, err := x509.ParseCertificateRequest(csrDER)
	if err != nil {
		return "", err
	}
	csr.PublicKey = pubKey
	// Encode lại CSR với public key mới
	finalCSRDER, err := x509.CreateCertificateRequest(rand.Reader, &tmpl, nil)
	if err != nil {
		return "", err
	}
	csrPEM = string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: finalCSRDER}))
	return csrPEM, nil
}

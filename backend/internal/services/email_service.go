package services

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/trustwired/internal-t/internal/config"
)

type EmailService struct {
	Config *config.Config
}

func NewEmailService(cfg *config.Config) *EmailService {
	return &EmailService{Config: cfg}
}

type ResendEmail struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

func (s *EmailService) Send(to, subject, html string) error {
	if s.Config.ResendAPIKey == "" {
		fmt.Printf("[EMAIL] To: %s | Subject: %s\n", to, subject)
		return nil
	}

	payload := ResendEmail{
		From:    fmt.Sprintf("%s <%s>", s.Config.MailFromName, s.Config.MailFromAddress),
		To:      []string{to},
		Subject: subject,
		HTML:    html,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.Config.ResendAPIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("resend API error %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func (s *EmailService) SendVerificationEmail(userID uint, email string) error {
	emailHash := sha1Hash(email)
	expires := time.Now().Add(60 * time.Minute).Unix()

	params := url.Values{}
	params.Set("id", strconv.FormatUint(uint64(userID), 10))
	params.Set("hash", emailHash)
	params.Set("expires", strconv.FormatInt(expires, 10))
	sig := s.generateSignature(params)
	params.Set("signature", sig)

	verifyURL := fmt.Sprintf("%s/verify-email?%s", s.Config.FrontendURL, params.Encode())

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<body style="font-family: sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
  <h2>Verify Your Email Address</h2>
  <p>Welcome to Trustwired Internal System. Please click the button below to verify your email address.</p>
  <a href="%s" style="display: inline-block; background: #2563eb; color: white; padding: 12px 24px; border-radius: 6px; text-decoration: none; margin: 20px 0;">
    Verify Email
  </a>
  <p>This link expires in 60 minutes.</p>
  <p>If you did not create an account, no further action is required.</p>
</body>
</html>`, verifyURL)

	return s.Send(email, "Verify your email address - Trustwired", html)
}

func (s *EmailService) SendWelcomeEmail(name, email, tempPassword string) error {
	loginURL := s.Config.FrontendURL + "/login"

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<body style="font-family: sans-serif; max-width: 600px; margin: 0 auto; padding: 40px 20px; background: #f9fafb;">
  <div style="background: white; border-radius: 12px; padding: 40px; box-shadow: 0 1px 3px rgba(0,0,0,0.1);">
    <h2 style="margin: 0 0 8px; color: #111827;">Welcome to Trustwired, %s!</h2>
    <p style="color: #6b7280; margin: 0 0 24px;">Your account has been created. Use the credentials below to sign in — you will be asked to set a new password immediately after.</p>
    <div style="background: #f9fafb; border: 1px solid #e5e7eb; border-radius: 8px; padding: 20px; margin-bottom: 24px;">
      <p style="margin: 0 0 8px; font-size: 14px; color: #374151;"><strong>Email:</strong> %s</p>
      <p style="margin: 0; font-size: 14px; color: #374151;"><strong>Temporary password:</strong> <code style="background:#e5e7eb;padding:2px 6px;border-radius:4px;">%s</code></p>
    </div>
    <a href="%s" style="display: inline-block; background: #4f46e5; color: white; padding: 14px 28px; border-radius: 8px; text-decoration: none; font-weight: 600; font-size: 15px;">
      Sign In Now
    </a>
    <p style="color: #9ca3af; font-size: 13px; margin: 32px 0 0;">If you weren't expecting this email, please contact your administrator.</p>
  </div>
</body>
</html>`, name, email, tempPassword, loginURL)

	return s.Send(email, "Welcome to Trustwired — your account is ready", html)
}

func (s *EmailService) SendMagicLinkEmail(email, token string) error {
	magicURL := fmt.Sprintf("%s/magic-link/verify?token=%s&email=%s",
		s.Config.FrontendURL,
		url.QueryEscape(token),
		url.QueryEscape(email),
	)

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<body style="font-family: sans-serif; max-width: 600px; margin: 0 auto; padding: 40px 20px; background: #f9fafb;">
  <div style="background: white; border-radius: 12px; padding: 40px; box-shadow: 0 1px 3px rgba(0,0,0,0.1);">
    <h2 style="margin: 0 0 8px; color: #111827; font-size: 22px;">Your sign-in link</h2>
    <p style="color: #6b7280; margin: 0 0 32px;">Click the button below to sign in to Trustwired. This link expires in <strong>15 minutes</strong> and can only be used once.</p>
    <a href="%s" style="display: inline-block; background: #4f46e5; color: white; padding: 14px 28px; border-radius: 8px; text-decoration: none; font-weight: 600; font-size: 15px;">
      Sign in to Trustwired
    </a>
    <p style="color: #9ca3af; font-size: 13px; margin: 32px 0 0;">If you didn't request this link, you can safely ignore this email.</p>
    <hr style="border: none; border-top: 1px solid #f3f4f6; margin: 24px 0;" />
    <p style="color: #9ca3af; font-size: 12px; margin: 0;">Or copy this URL into your browser:<br>
    <span style="color: #6b7280; word-break: break-all;">%s</span></p>
  </div>
</body>
</html>`, magicURL, magicURL)

	return s.Send(email, "Your sign-in link — Trustwired", html)
}

func (s *EmailService) SendPasswordResetEmail(email, token string) error {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s&email=%s",
		s.Config.FrontendURL,
		url.QueryEscape(token),
		url.QueryEscape(email),
	)

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<body style="font-family: sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
  <h2>Reset Your Password</h2>
  <p>You are receiving this email because we received a password reset request for your account.</p>
  <a href="%s" style="display: inline-block; background: #2563eb; color: white; padding: 12px 24px; border-radius: 6px; text-decoration: none; margin: 20px 0;">
    Reset Password
  </a>
  <p>This password reset link will expire in 60 minutes.</p>
  <p>If you did not request a password reset, no further action is required.</p>
</body>
</html>`, resetURL)

	return s.Send(email, "Reset your password - Trustwired", html)
}

func (s *EmailService) VerifySignedURL(userID uint, email, expiresStr, signature string) bool {
	expires, err := strconv.ParseInt(expiresStr, 10, 64)
	if err != nil {
		return false
	}

	if time.Now().Unix() > expires {
		return false
	}

	params := url.Values{}
	params.Set("id", strconv.FormatUint(uint64(userID), 10))
	params.Set("hash", sha1Hash(email))
	params.Set("expires", expiresStr)

	expectedSig := s.generateSignature(params)
	return hmac.Equal([]byte(signature), []byte(expectedSig))
}

func (s *EmailService) generateSignature(params url.Values) string {
	message := params.Encode()
	mac := hmac.New(sha256.New, []byte(s.Config.AppKey))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

func sha1Hash(input string) string {
	h := sha1.New()
	h.Write([]byte(input))
	return hex.EncodeToString(h.Sum(nil))
}

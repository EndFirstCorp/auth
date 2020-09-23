package auth

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/pkg/errors"
)

func TestAuthStoreEndToEnd(t *testing.T) {
	b := NewBackendMemory(&hashStore{}).(*backendMemory)
	m := &TextMailer{}

	// Register emails
	verifyCode, err := _register("test@test.com", b, m)
	if err != nil {
		t.Fatal("Register:", err)
	}
	verifyCode2, err := _register("test2@test.com", b, m)
	if err != nil {
		t.Fatal("Register2", err)
	}

	// Verify emails
	csrfToken, emailCookie, err := _verify(verifyCode, b, m)
	if err != nil {
		t.Fatal("Verify:", err)
	}
	csrfToken2, emailCookie2, err := _verify(verifyCode2, b, m)
	if err != nil {
		t.Fatal("Verify2:", err)
	}

	// Create user Profile
	csrfToken, sessionCookie, err := _createProfile("fullName", "password", emailCookie, b, m, csrfToken)
	if err != nil {
		t.Fatal("User profile:", err)
	}
	csrfToken2, sessionCookie2, err := _createProfile("fullName2", "password2", emailCookie2, b, m, csrfToken2)
	if err != nil {
		t.Fatal("User profile2:", err)
	}

	// login and create remember me
	csrfToken, sessionCookie, rememberCookie, err := _login("test@test.com", "password", true, sessionCookie, nil, b, m)
	if err != nil {
		t.Fatal("Login:", err)
	}
	// login, no remember me
	csrfToken2, sessionCookie2, rememberCookie2, err := _login("test2@test.com", "password2", false, sessionCookie2, nil, b, m)
	if err != nil {
		t.Fatal("Login2:", err)
	}

	sessionHash, _ := decodeStringToHash(sessionCookie.SessionID)
	session := b.getSessionByHash(sessionHash)
	session2Hash, _ := decodeStringToHash(sessionCookie2.SessionID)
	session2 := b.getSessionByHash(session2Hash)
	// authenticate. Both should work
	err = _auth(sessionCookie, rememberCookie, b, m, csrfToken)
	if err != nil {
		t.Fatal("Auth:", err)
	}
	err = _auth(sessionCookie2, rememberCookie2, b, m, csrfToken2)
	if err != nil {
		t.Fatal("Auth:", err)
	}

	// renews in 2 seconds
	session.RenewTimeUTC = time.Now().UTC().Add(2 * time.Second)
	err = _auth(sessionCookie, rememberCookie, b, m, csrfToken)
	if err != nil {
		t.Fatal("Auth:", err)
	}
	// authenticate, expired session, but should renew due to rememberMe
	session.ExpireTimeUTC = time.Now().AddDate(0, 0, -1)
	err = _auth(sessionCookie, rememberCookie, b, m, csrfToken)
	if err != nil {
		t.Fatal("Auth:", err)
	}
	// non-expired, but requires renewal
	session2.RenewTimeUTC = time.Now().Add(-5 * time.Minute)
	err = _auth(sessionCookie2, rememberCookie2, b, m, csrfToken2)
	if err != nil {
		t.Fatal("Auth:", err)
	}
	// expired session, but can't renew due to no remember me
	session2.ExpireTimeUTC = time.Now().AddDate(0, 0, -1)
	err = _auth(sessionCookie2, rememberCookie2, b, m, csrfToken2)
	if err == nil || err.Error() != "Unable to renew session" {
		t.Fatal("Auth:", err)
	}
}

func _register(email string, b *backendMemory, m *TextMailer) (string, error) {
	r := &http.Request{Header: http.Header{}}
	c := NewMockCookieStore(nil, false, false)
	s := &authStore{b, m, c}
	lenSessions := len(b.EmailSessions)

	// register new user
	// adds to users, logins and sessions
	err := s.register(r, b, EmailSendParams{Email: email, TemplateSuccess: "templateName", SubjectSuccess: "emailSubject", Info: map[string]interface{}{"key": "value"}}, "")
	if err != nil {
		return "", err
	}
	if len(b.EmailSessions) != lenSessions+1 {
		return "", errors.New("expected to add a new Email session")
	}
	// get code from "email"
	data := m.MessageData.(EmailSendParams)
	emailVerifyHash, _ := decodeStringToHash(data.VerificationCode + "=")
	if b.EmailSessions[lenSessions].Email != email || b.EmailSessions[lenSessions].EmailVerifyHash != emailVerifyHash || b.EmailSessions[lenSessions].Info == nil || b.EmailSessions[lenSessions].Info["key"] != "value" {
		return "", errors.Errorf("expected to have valid session: %s, %v, %v", b.EmailSessions[lenSessions].Email, b.EmailSessions[lenSessions].EmailVerifyHash != emailVerifyHash, b.EmailSessions[lenSessions].Info != nil && b.EmailSessions[lenSessions].Info["key"] != "value")
	}

	return data.VerificationCode, nil
}

func _verify(verifyCode string, b *backendMemory, m *TextMailer) (string, *emailCookie, error) {
	r := &http.Request{Header: http.Header{}}
	c := NewMockCookieStore(nil, false, false)
	s := &authStore{b, m, c}
	lenEmailSessions := len(b.EmailSessions)
	lenUsers := len(b.Users)
	emailVerifyHash, _ := decodeStringToHash(verifyCode + "=")
	emailSession := b.getEmailSessionByEmailVerifyHash(emailVerifyHash)

	// verify Email. Should 1. add user to b.Users, 2. set UserID in EmailSession, 3. add session
	csrfToken, user, err := s.verifyEmail(nil, r, b, EmailSendParams{VerificationCode: verifyCode, TemplateSuccess: "templateName", SubjectSuccess: "emailSubject"})
	if err != nil {
		return "", nil, err
	}
	if user == nil || user.Info == nil || user.Info["key"] != "value" {
		return "", nil, errors.Errorf("expected to get back info that we entered during register phase")
	}
	if len(b.Users) != +lenUsers+1 || len(b.EmailSessions) != lenEmailSessions {
		return "", nil, errors.Errorf("expected to add user and update existing session: %v, %v", len(b.Users) != lenUsers+1, len(b.EmailSessions) != lenEmailSessions)
	}
	if b.Users[lenUsers].UserID != strconv.Itoa(b.LastUserID) || emailSession == nil || b.Users[lenUsers].PrimaryEmail != emailSession.Email {
		return "", nil, errors.Errorf("expected user to be added with new UserID and correct email: %v, %s, %s", b.Users[lenUsers].UserID != strconv.Itoa(b.LastUserID), b.Users[lenUsers].PrimaryEmail, emailSession.Email)
	}

	emailCookie, ok := c.cookies["Email"].(*emailCookie)
	if !ok {
		return "", nil, nil
	}
	if emailCookie.EmailVerificationCode != verifyCode+"=" {
		return "", nil, errors.Errorf("expected cookie Code to be correct: %s, %s", emailCookie.EmailVerificationCode, verifyCode+"=")
	}

	return csrfToken, emailCookie, nil
}

func _createProfile(fullName, password string, emailCookie *emailCookie, b *backendMemory, m *TextMailer, csrfToken string) (string, *sessionCookie, error) {
	r := &http.Request{Header: http.Header{}}
	c := NewMockCookieStore(map[string]interface{}{"Email": emailCookie}, false, false)
	s := &authStore{b, m, c}
	p := profile{Password: password}
	emailVerifyHash, _ := decodeStringToHash(emailCookie.EmailVerificationCode)
	oldEmailSession := b.getEmailSessionByEmailVerifyHash(emailVerifyHash)
	var user *user
	if oldEmailSession != nil {
		user = b.getUserByEmail(oldEmailSession.Email)
	}

	// create profile
	newSession, err := s.createProfile(nil, r, b, csrfToken, &p)
	if err != nil {
		return "", nil, err
	}
	// check password was saved correctly
	h := &hashStore{}
	passwordHash, err := h.Hash(password)
	if err != nil {
		return "", nil, err
	}
	// check user was saved correctly
	if user == nil || oldEmailSession == nil || user.PrimaryEmail != oldEmailSession.Email || user.UserID != oldEmailSession.UserID || user.PasswordHash != passwordHash || user.Info == nil || user.Info["key"] != "value" {
		return "", nil, errors.Errorf("expected user to be updated with expected values: %v, %v", user, oldEmailSession)
	}
	// verify email session was deleted
	if emailSession := b.getEmailSessionByEmailVerifyHash(emailVerifyHash); emailSession != nil {
		return "", nil, errors.Errorf("expected Email session to be deleted: %v", emailSession)
	}

	// verify session cookie
	sessionCookie := c.cookies["Session"].(*sessionCookie)
	sessionHash, _ := decodeStringToHash(sessionCookie.SessionID)
	session := b.getSessionByHash(sessionHash)
	if session == nil || session.SessionHash != sessionHash || session.Email != oldEmailSession.Email || session.UserID != oldEmailSession.UserID || session.Info == nil || session.Info["key"] != "value" {
		return "", nil, errors.Errorf("expected session to be created, %v", session)
	}
	return newSession.CSRFToken, sessionCookie, nil
}

func _login(email, password string, remember bool, clientSessionCookie *sessionCookie, rememberCookie *rememberMeCookie, b *backendMemory, m *TextMailer) (string, *sessionCookie, *rememberMeCookie, error) {
	r := &http.Request{Header: http.Header{}}
	c := NewMockCookieStore(map[string]interface{}{"Session": clientSessionCookie, "RememberMe": rememberCookie}, false, false)
	s := &authStore{b, m, c}
	lenUsers := len(b.Users)

	// login
	session, err := s.login(nil, r, b, email, password, remember)
	if err != nil {
		return "", nil, nil, err
	}
	// verify session is valid
	user := b.getUserByEmail(email)
	if session == nil || session.Email != email || session.UserID != user.UserID || session.Info == nil || session.Info["key"] != "value" {
		return "", nil, nil, errors.Errorf("session wasn't created correctly, %v", session)
	}
	// verify no users were created
	if lenUsers != len(b.Users) {
		return "", nil, nil, errors.Errorf("expected no new users to be created, %v", lenUsers != len(b.Users))
	}
	// verify old session and old remember me were deleted
	if clientSessionCookie != nil {
		oldSessionHash, _ := decodeStringToHash(clientSessionCookie.SessionID)
		oldSession := b.getSessionByHash(oldSessionHash)
		if oldSession != nil {
			return "", nil, nil, errors.Errorf("expected old session to be deleted")
		}
	}
	if rememberCookie != nil {
		oldRemember := b.getRememberMe(rememberCookie.Selector)
		if oldRemember != nil {
			return "", nil, nil, errors.Errorf("expected old remember me to be deleted")
		}
	}

	// get and verify session cookie
	newSessionCookie := c.cookies["Session"].(*sessionCookie)
	sessionHash, _ := decodeStringToHash(newSessionCookie.SessionID)
	newSession := b.getSessionByHash(sessionHash)
	if newSession == nil || newSession.SessionHash != sessionHash || newSession.Email != session.Email || newSession.UserID != session.UserID || newSession.Info == nil || newSession.Info["key"] != "value" {
		return "", nil, nil, errors.Errorf("expected session to be created in database that matches return from function, %v", newSession)
	}
	// verify rememberMe cookie
	newRememberCookie, ok := c.cookies["RememberMe"].(*rememberMeCookie)
	if remember {
		if !ok {
			return "", nil, nil, errors.New("expected rememberMe cookie to be created")
		}
		tokenHash, _ := decodeStringToHash(newRememberCookie.Token)
		rememberMe := b.getRememberMe(newRememberCookie.Selector)
		if rememberMe == nil || rememberMe.Selector != newRememberCookie.Selector || rememberMe.Email != email || rememberMe.UserID != user.UserID || rememberMe.TokenHash != tokenHash {
			return "", nil, nil, errors.Errorf("expected valid remember me to be created, %v", rememberMe)
		}
	} else if newRememberCookie != nil {
		return "", nil, nil, errors.Errorf("unexpected remember cookie when not supposed to create one: %v", newRememberCookie)
	}

	return session.CSRFToken, newSessionCookie, newRememberCookie, nil
}

func _auth(clientSessionCookie *sessionCookie, rememberCookie *rememberMeCookie, b *backendMemory, m *TextMailer, csrfToken string) error {
	r := &http.Request{Header: http.Header{}}
	r.Header.Add("X-CSRF-Token", csrfToken)
	c := NewMockCookieStore(map[string]interface{}{"Session": clientSessionCookie, "RememberMe": rememberCookie}, false, false)
	s := &authStore{b, m, c}
	session, err := s.GetSession(nil, r)
	if err != nil {
		return err
	}
	if clientSessionCookie != nil {
		sessionHash, _ := decodeStringToHash(clientSessionCookie.SessionID)
		if session.SessionHash != sessionHash {
			return errors.Errorf("expected to maintain sessionHash: %s, %s", sessionHash, session.SessionHash)
		}
	}
	sc := c.cookies["Session"].(*sessionCookie)
	if sc != nil {
		if sc.ExpireTimeUTC.Sub(time.Now()) > sessionExpireDuration || sc.ExpireTimeUTC.Sub(time.Now()) < 0 {
			return errors.New("expire time should be between 0 and sessionExpireDuration")
		}
		if sc.RenewTimeUTC.Sub(time.Now()) > sessionRenewDuration || sc.RenewTimeUTC.Sub(time.Now()) < 0 {
			return errors.New("expire time should be between 0 and sessionRenewDuration")
		}
	}
	rc := c.cookies["RememberMe"].(*rememberMeCookie)
	if rc != nil {
		rememberMe := b.getRememberMe(rc.Selector)
		if rememberMe == nil || rememberMe.ExpireTimeUTC != rc.ExpireTimeUTC || rememberMe.RenewTimeUTC != rc.RenewTimeUTC {
			return errors.New("expected valid rememberMe in database if remember cookie is set")
		}
		if rememberMe.ExpireTimeUTC.Sub(sc.ExpireTimeUTC) < 0 || rememberMe.ExpireTimeUTC.Sub(sc.RenewTimeUTC) < 0 {
			return errors.New("session expire and renew time cannot exceed rememberMe expire")
		}
	}
	// code attempted to do a renew
	if clientSessionCookie.RenewTimeUTC.Before(time.Now().UTC()) || clientSessionCookie.ExpireTimeUTC.Before(time.Now().UTC()) {

	}
	return nil
}

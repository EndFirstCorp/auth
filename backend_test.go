package main

import (
	"errors"
	"testing"
	"time"
)

func TestAuthError(t *testing.T) {
	e3 := newAuthError("error 3", errors.New("other"))
	e2 := newAuthError("error 2", e3)
	e1 := newAuthError("error 1", e2)
	if e1.message != "error 1" || e2.message != "error 2" || e1.Error() != e1.message || e2.Error() != e2.message ||
		e1.Trace() != "error 1\n  error 2\n    error 3\n      other\n" ||
		e2.Trace() != "error 2\n  error 3\n    other\n" ||
		e3.Trace() != "error 3\n  other\n" {
		t.Error("expected valid error structs", e1, e2, e3, e1.Trace(), e2.Trace(), e3.Trace())
	}
}

func TestBackendLogin(t *testing.T) {
	m := &mockBackend{LoginReturn: loginSuccess()}
	b := backend{u: m, l: m, s: m}
	b.Login("email", "password")
	if len(m.MethodsCalled) != 1 || m.MethodsCalled[0] != "Login" {
		t.Error("Expected it would call backend", m.MethodsCalled)
	}
}

func TestBackendCreateSession(t *testing.T) {
	m := &mockBackend{CreateSessionReturn: sessionRemember(time.Now(), time.Now())}
	b := backend{u: m, l: m, s: m}
	b.CreateSession(1, "test@test.com", "fullname", "hash", time.Now(), time.Now(), false, "", "", time.Now(), time.Now())
	if len(m.MethodsCalled) != 1 || m.MethodsCalled[0] != "CreateSession" {
		t.Error("Expected it would call backend", m.MethodsCalled)
	}
}

func TestBackendGetSession(t *testing.T) {
	m := &mockBackend{GetSessionReturn: sessionErr()}
	b := backend{u: m, l: m, s: m}
	b.GetSession("hash")
	if len(m.MethodsCalled) != 1 || m.MethodsCalled[0] != "GetSession" {
		t.Error("Expected it would call backend", m.MethodsCalled)
	}
}

func TestBackendRenewSession(t *testing.T) {
	m := &mockBackend{RenewSessionReturn: sessionErr()}
	b := backend{u: m, l: m, s: m}
	b.RenewSession("hash", time.Now())
	if len(m.MethodsCalled) != 1 || m.MethodsCalled[0] != "RenewSession" {
		t.Error("Expected it would call backend", m.MethodsCalled)
	}
}

func TestBackendGetRememberMe(t *testing.T) {
	m := &mockBackend{GetRememberMeReturn: rememberErr()}
	b := backend{u: m, l: m, s: m}
	b.GetRememberMe("selector")
	if len(m.MethodsCalled) != 1 || m.MethodsCalled[0] != "GetRememberMe" {
		t.Error("Expected it would call backend", m.MethodsCalled)
	}
}

func TestBackendRenewRememberMe(t *testing.T) {
	m := &mockBackend{RenewRememberMeReturn: rememberErr()}
	b := backend{u: m, l: m, s: m}
	b.RenewRememberMe("selector", time.Now())
	if len(m.MethodsCalled) != 1 || m.MethodsCalled[0] != "RenewRememberMe" {
		t.Error("Expected it would call backend", m.MethodsCalled)
	}
}

func TestBackendAddUser(t *testing.T) {
	m := &mockBackend{AddUserReturn: nil}
	b := backend{u: m, l: m, s: m}
	b.AddUser("mail")
	if len(m.MethodsCalled) != 1 || m.MethodsCalled[0] != "AddUser" {
		t.Error("Expected it would call backend", m.MethodsCalled)
	}
}

func TestBackendGetEmailSession(t *testing.T) {
	m := &mockBackend{getEmailSessionReturn: getEmailSessionErr()}
	b := backend{u: m, l: m, s: m}
	b.GetEmailSession("hash")
	if len(m.MethodsCalled) != 1 || m.MethodsCalled[0] != "GetEmailSession" {
		t.Error("Expected it would call backend", m.MethodsCalled)
	}
}

func TestBackendUpdateUser(t *testing.T) {
	m := &mockBackend{}
	b := backend{u: m, l: m, s: m}
	b.UpdateUser("hash", "name", "company", "url")
	if len(m.MethodsCalled) != 1 || m.MethodsCalled[0] != "UpdateUser" {
		t.Error("Expected it would call backend", m.MethodsCalled)
	}
}

func TestBackendCreateLogin(t *testing.T) {
	m := &mockBackend{CreateLoginReturn: loginErr()}
	b := backend{u: m, l: m, s: m}
	b.CreateLogin(1, "email", "hash", "name", "homeDir", 1, 1, "quota", "fileQuota")
	if len(m.MethodsCalled) != 1 || m.MethodsCalled[0] != "CreateLogin" {
		t.Error("Expected it would call backend", m.MethodsCalled)
	}
}

func TestBackendUpdateEmail(t *testing.T) {
	m := &mockBackend{UpdateEmailReturn: sessionErr()}
	b := backend{u: m, l: m, s: m}
	b.UpdateEmail("email", "password", "newEmail")
	if len(m.MethodsCalled) != 1 || m.MethodsCalled[0] != "UpdateEmail" {
		t.Error("Expected it would call backend", m.MethodsCalled)
	}
}

func TestBackendUpdatePassword(t *testing.T) {
	m := &mockBackend{UpdatePasswordReturn: sessionErr()}
	b := backend{u: m, l: m, s: m}
	b.UpdatePassword("email", "oldPassword", "newPassword")
	if len(m.MethodsCalled) != 1 || m.MethodsCalled[0] != "UpdatePassword" {
		t.Error("Expected it would call backend", m.MethodsCalled)
	}
}

func TestBackendInvalidateSession(t *testing.T) {
	m := &mockBackend{}
	b := backend{u: m, l: m, s: m}
	b.InvalidateSession("hash")
	if len(m.MethodsCalled) != 1 || m.MethodsCalled[0] != "InvalidateSession" {
		t.Error("Expected it would call backend", m.MethodsCalled)
	}
}

func TestBackendInvalidateSessions(t *testing.T) {
	m := &mockBackend{}
	b := backend{u: m, l: m, s: m}
	b.InvalidateSessions("email")
	if len(m.MethodsCalled) != 1 || m.MethodsCalled[0] != "InvalidateSessions" {
		t.Error("Expected it would call backend", m.MethodsCalled)
	}
}

func TestBackendInvalidateRememberMe(t *testing.T) {
	m := &mockBackend{}
	b := backend{u: m, l: m, s: m}
	b.InvalidateRememberMe("selector")
	if len(m.MethodsCalled) != 1 || m.MethodsCalled[0] != "InvalidateRememberMe" {
		t.Error("Expected it would call backend", m.MethodsCalled)
	}
}

func TestBackendClose(t *testing.T) {
	// all succeed
	m := &mockBackend{}
	b := backend{u: m, l: m, s: m}
	b.Close()
	if len(m.MethodsCalled) != 3 || m.MethodsCalled[0] != "Close" || m.MethodsCalled[1] != "Close" || m.MethodsCalled[2] != "Close" {
		t.Error("Expected it would call backend", m.MethodsCalled)
	}

	// error on session close
	m = &mockBackend{}
	e := &mockBackend{ErrReturn: errors.New("failed")}
	b = backend{u: m, l: m, s: e}
	b.Close()
	if len(m.MethodsCalled) != 0 || len(e.MethodsCalled) != 1 || e.MethodsCalled[0] != "Close" {
		t.Error("Expected fail on session close", m.MethodsCalled)
	}

	// error on user close
	m = &mockBackend{}
	e = &mockBackend{ErrReturn: errors.New("failed")}
	b = backend{u: e, l: m, s: m}
	b.Close()
	if len(e.MethodsCalled) != 1 || len(m.MethodsCalled) != 1 || m.MethodsCalled[0] != "Close" || e.MethodsCalled[0] != "Close" {
		t.Error("Expected fail on user close", m.MethodsCalled)
	}

	// error on login close
	m = &mockBackend{}
	e = &mockBackend{ErrReturn: errors.New("failed")}
	b = backend{u: m, l: e, s: m}
	b.Close()
	if len(m.MethodsCalled) != 2 || len(e.MethodsCalled) != 1 || m.MethodsCalled[0] != "Close" || m.MethodsCalled[1] != "Close" || e.MethodsCalled[0] != "Close" {
		t.Error("Expected it would call backend", m.MethodsCalled)
	}
}

/***********************************************************************/

type LoginReturn struct {
	Login *userLogin
	Err   error
}

type SessionReturn struct {
	Session *loginSession
	Err     error
}

type SessionRememberReturn struct {
	Session    *loginSession
	RememberMe *rememberMeSession
	Err        error
}

type RememberMeReturn struct {
	RememberMe *rememberMeSession
	Err        error
}

type GetUserReturn struct {
	User *user
	Err  error
}

type getEmailSessionReturn struct {
	Session *emailSession
	Err     error
}

type mockBackend struct {
	backender
	LoginReturn              *LoginReturn
	ExpirationReturn         *time.Time
	GetSessionReturn         *SessionReturn
	CreateSessionReturn      *SessionRememberReturn
	RenewSessionReturn       *SessionReturn
	AddUserReturn            error
	DeleteEmailSessionReturn error
	UpdateEmailSessionReturn error
	GetUserReturn            *GetUserReturn
	getEmailSessionReturn    *getEmailSessionReturn
	CreateLoginReturn        *LoginReturn
	UpdateEmailReturn        *SessionReturn
	UpdatePasswordReturn     *SessionReturn
	GetRememberMeReturn      *RememberMeReturn
	RenewRememberMeReturn    *RememberMeReturn
	RememberMeReturn         *RememberMeReturn
	ErrReturn                error
	MethodsCalled            []string
}

func (b *mockBackend) Login(email, password string) (*userLogin, error) {
	b.MethodsCalled = append(b.MethodsCalled, "Login")
	if b.LoginReturn == nil {
		return nil, errors.New("LoginReturn not initialized")
	}
	return b.LoginReturn.Login, b.LoginReturn.Err
}

func (b *mockBackend) GetSession(sessionHash string) (*loginSession, error) {
	b.MethodsCalled = append(b.MethodsCalled, "GetSession")
	if b.GetSessionReturn == nil {
		return nil, errors.New("GetSessionReturn not initialized")
	}
	return b.GetSessionReturn.Session, b.GetSessionReturn.Err
}

func (b *mockBackend) CreateSession(userID int, email, fullname, sessionHash string, sessionRenewTimeUTC, sessionExpireTimeUTC time.Time, rememberMe bool, rememberMeSelector, rememberMeTokenHash string, rememberMeRenewTimeUTC, rememberMeExpireTimeUTC time.Time) (*loginSession, *rememberMeSession, error) {
	b.MethodsCalled = append(b.MethodsCalled, "CreateSession")
	if b.CreateSessionReturn == nil {
		return nil, nil, errors.New("CreateSessionReturn not initialized")
	}
	return b.CreateSessionReturn.Session, b.CreateSessionReturn.RememberMe, b.CreateSessionReturn.Err
}

func (b *mockBackend) RenewSession(sessionHash string, renewTimeUTC time.Time) (*loginSession, error) {
	b.MethodsCalled = append(b.MethodsCalled, "RenewSession")
	if b.RenewSessionReturn == nil {
		return nil, errors.New("RenewSessionReturn not initialized")
	}
	return b.RenewSessionReturn.Session, b.RenewSessionReturn.Err
}
func (b *mockBackend) GetRememberMe(selector string) (*rememberMeSession, error) {
	b.MethodsCalled = append(b.MethodsCalled, "GetRememberMe")
	if b.GetRememberMeReturn == nil {
		return nil, errors.New("GetRememberMeReturn not initialized")
	}
	return b.GetRememberMeReturn.RememberMe, b.GetRememberMeReturn.Err
}
func (b *mockBackend) RenewRememberMe(selector string, renewTimeUTC time.Time) (*rememberMeSession, error) {
	b.MethodsCalled = append(b.MethodsCalled, "RenewRememberMe")
	if b.RenewRememberMeReturn == nil {
		return nil, errors.New("RenewRememberMeReturn not initialized")
	}
	return b.RenewRememberMeReturn.RememberMe, b.RenewRememberMeReturn.Err
}
func (b *mockBackend) AddUser(email string) (int, error) {
	b.MethodsCalled = append(b.MethodsCalled, "AddUser")
	return 1, b.AddUserReturn
}

func (b *mockBackend) CreateEmailSession(email, emailVerifyHash string) error {
	b.MethodsCalled = append(b.MethodsCalled, "CreateEmailSession")
	return b.ErrReturn
}

func (b *mockBackend) GetEmailSession(emailVerifyHash string) (*emailSession, error) {
	b.MethodsCalled = append(b.MethodsCalled, "GetEmailSession")
	if b.getEmailSessionReturn == nil {
		return nil, errors.New("getEmailSessionReturn not initialized")
	}
	return b.getEmailSessionReturn.Session, b.getEmailSessionReturn.Err
}

func (b *mockBackend) UpdateEmailSession(emailVerifyHash string, userID int, email string) error {
	b.MethodsCalled = append(b.MethodsCalled, "UpdateEmailSession")
	return b.UpdateEmailSessionReturn
}

func (b *mockBackend) DeleteEmailSession(emailVerifyHash string) error {
	b.MethodsCalled = append(b.MethodsCalled, "DeleteEmailSession")
	return b.DeleteEmailSessionReturn
}

func (b *mockBackend) GetUser(email string) (*user, error) {
	b.MethodsCalled = append(b.MethodsCalled, "GetUser")
	if b.GetUserReturn == nil {
		return nil, errors.New("GetUserReturn not initialized")
	}
	return b.GetUserReturn.User, b.GetUserReturn.Err
}

func (b *mockBackend) UpdateUser(email, fullname, company, pictureURL string) error {
	b.MethodsCalled = append(b.MethodsCalled, "UpdateUser")
	return b.ErrReturn
}

func (b *mockBackend) CreateLogin(userID int, email, passwordHash, fullName, homeDirectory string, uidNumber, gidNumber int, mailQuota, fileQuota string) (*userLogin, error) {
	b.MethodsCalled = append(b.MethodsCalled, "CreateLogin")
	if b.CreateLoginReturn == nil {
		return nil, errors.New("CreateLoginReturn not initialized")
	}
	return b.CreateLoginReturn.Login, b.CreateLoginReturn.Err
}

func (b *mockBackend) UpdateEmail(email string, password string, newEmail string) (*loginSession, error) {
	b.MethodsCalled = append(b.MethodsCalled, "UpdateEmail")
	if b.UpdateEmailReturn == nil {
		return nil, errors.New("UpdateEmailReturn not initialized")
	}
	return b.UpdateEmailReturn.Session, b.UpdateEmailReturn.Err
}

func (b *mockBackend) UpdatePassword(email string, oldPassword string, newPassword string) (*loginSession, error) {
	b.MethodsCalled = append(b.MethodsCalled, "UpdatePassword")
	if b.UpdatePasswordReturn == nil {
		return nil, errors.New("UpdatePasswordReturn not initialized")
	}
	return b.UpdatePasswordReturn.Session, b.UpdatePasswordReturn.Err
}

func (b *mockBackend) InvalidateSession(sessionHash string) error {
	b.MethodsCalled = append(b.MethodsCalled, "InvalidateSession")
	return b.ErrReturn
}

func (b *mockBackend) InvalidateSessions(email string) error {
	b.MethodsCalled = append(b.MethodsCalled, "InvalidateSessions")
	return b.ErrReturn
}

func (b *mockBackend) InvalidateRememberMe(selector string) error {
	b.MethodsCalled = append(b.MethodsCalled, "InvalidateRememberMe")
	return b.ErrReturn
}

func (b *mockBackend) Close() error {
	b.MethodsCalled = append(b.MethodsCalled, "Close")
	return b.ErrReturn
}

func loginSuccess() *LoginReturn {
	return &LoginReturn{&userLogin{Email: "test@test.com"}, nil}
}

func loginErr() *LoginReturn {
	return &LoginReturn{nil, errors.New("failed")}
}

func sessionSuccess(renewTimeUTC, expireTimeUTC time.Time) *SessionReturn {
	return &SessionReturn{&loginSession{1, "test@test.com", "fullname", "sessionHash", renewTimeUTC, expireTimeUTC}, nil}
}

func sessionErr() *SessionReturn {
	return &SessionReturn{&loginSession{}, errors.New("failed")}
}

func rememberMe(renewTimeUTC, expireTimeUTC time.Time) *RememberMeReturn { // hash of the word "token"
	return &RememberMeReturn{&rememberMeSession{TokenHash: "PEaenWxYddN6Q_NT1PiOYfz4EsZu7jRXRlpAsNpBU-A=", ExpireTimeUTC: expireTimeUTC, RenewTimeUTC: renewTimeUTC}, nil}
}

func rememberErr() *RememberMeReturn {
	return &RememberMeReturn{&rememberMeSession{}, errors.New("failed")}
}

func sessionRemember(renewTimeUTC, expireTimeUTC time.Time) *SessionRememberReturn {
	return &SessionRememberReturn{&loginSession{1, "test@test.com", "fullname", "sessionHash", renewTimeUTC, expireTimeUTC}, &rememberMeSession{TokenHash: "PEaenWxYddN6Q_NT1PiOYfz4EsZu7jRXRlpAsNpBU-A=", ExpireTimeUTC: expireTimeUTC, RenewTimeUTC: renewTimeUTC}, nil}
}

func sessionRememberErr() *SessionRememberReturn {
	return &SessionRememberReturn{nil, nil, errors.New("failed")}
}

func getEmailSessionSuccess() *getEmailSessionReturn {
	return &getEmailSessionReturn{&emailSession{Email: "email", EmailVerifyHash: "hash"}, nil}
}
func getEmailSessionErr() *getEmailSessionReturn {
	return &getEmailSessionReturn{nil, errors.New("failed")}
}

func getUserSuccess() *GetUserReturn {
	return &GetUserReturn{&user{FullName: "name", PrimaryEmail: "test@test.com"}, nil}
}

func getUserErr() *GetUserReturn {
	return &GetUserReturn{nil, errors.New("failed")}
}

package main

import (
	"net/http"
	"testing"
	"time"
)

func getSessionStore(emailCookieToReturn *EmailCookie, sessionCookieToReturn *SessionCookie, rememberMeCookieToReturn *RememberMeCookie, hasCookieGetError, hasCookiePutError bool, backend *MockBackend) *SessionStore {
	r := &http.Request{}
	cookieStore := NewMockCookieStore(map[string]interface{}{emailCookieName: emailCookieToReturn, sessionCookieName: sessionCookieToReturn, rememberMeCookieName: rememberMeCookieToReturn}, hasCookieGetError, hasCookiePutError)
	return &SessionStore{backend, cookieStore, r}
}

var getSessionTests = []struct {
	Scenario            string
	HasCookieGetError   bool
	HasCookiePutError   bool
	SessionCookie       *SessionCookie
	GetSessionReturn    *SessionReturn
	RenewSessionReturn  *SessionReturn
	GetRememberMeReturn *RememberMeReturn
	MethodsCalled       []string
	ExpectedResult      *UserLoginRememberMe
	ExpectedErr         string
}{
	{
		Scenario:         "Get Session Valid",
		SessionCookie:    sessionCookie(futureTime, futureTime),
		GetSessionReturn: sessionSuccess(futureTime, futureTime),
		MethodsCalled:    []string{"GetSession"},
	},
	{
		Scenario:          "Get Session Cookie Error",
		HasCookieGetError: true,
		ExpectedErr:       "Session cookie not found",
	},
	{
		Scenario:      "Get Session Invalid Cookie Error",
		SessionCookie: sessionBogusCookie(futureTime, futureTime),
		ExpectedErr:   "Unable to decode session cookie",
	},
	{
		Scenario:         "Get Session Error",
		SessionCookie:    sessionCookie(futureTime, futureTime),
		GetSessionReturn: &SessionReturn{&UserLoginSession{}, ErrSessionNotFound},
		MethodsCalled:    []string{"GetSession"},
		ExpectedErr:      "Failed to verify session",
	},
	{
		Scenario:           "Get Session Renew",
		SessionCookie:      sessionCookie(pastTime, futureTime),
		RenewSessionReturn: sessionSuccess(futureTime, futureTime),
		MethodsCalled:      []string{"RenewSession"},
	},
}

func TestGetSession(t *testing.T) {
	for i, test := range getSessionTests {
		backend := &MockBackend{GetSessionReturn: test.GetSessionReturn, RenewSessionReturn: test.RenewSessionReturn}
		store := getSessionStore(nil, test.SessionCookie, nil, test.HasCookieGetError, test.HasCookiePutError, backend)
		val, err := store.GetSession()
		methods := store.backend.(*MockBackend).MethodsCalled
		if (err == nil && test.ExpectedErr != "" || err != nil && test.ExpectedErr != err.Error()) ||
			!collectionEqual(test.MethodsCalled, methods) {
			t.Errorf("Scenario[%d] failed: %s\nexpected err:%v\tactual err:%v\nexpected val:%v\tactual val:%v\nexpected methods: %s\tactual methods: %s", i, test.Scenario, test.ExpectedErr, err, test.ExpectedResult, val, test.MethodsCalled, methods)
		}
	}
}

var renewSessionTests = []struct {
	Scenario            string
	RenewTimeUTC        time.Time
	ExpireTimeUTC       time.Time
	HasCookieGetError   bool
	HasCookiePutError   bool
	RememberCookie      *RememberMeCookie
	RenewSessionReturn  *SessionReturn
	GetRememberMeReturn *RememberMeReturn
	MethodsCalled       []string
	ExpectedResult      *UserLoginRememberMe
	ExpectedErr         string
}{
	{
		Scenario:           "Renew Error",
		RenewTimeUTC:       pastTime,
		ExpireTimeUTC:      futureTime,
		RenewSessionReturn: sessionErr(),
		MethodsCalled:      []string{"RenewSession"},
		ExpectedErr:        "Unable to renew session",
	},
	{
		Scenario:           "Renew Save cookie error",
		RenewTimeUTC:       pastTime,
		ExpireTimeUTC:      futureTime,
		HasCookiePutError:  true,
		RenewSessionReturn: sessionSuccess(futureTime, futureTime),
		MethodsCalled:      []string{"RenewSession"},
		ExpectedErr:        "Error saving session cookie",
	},
	{
		Scenario:          "Error Getting RememberMe",
		RenewTimeUTC:      pastTime,
		ExpireTimeUTC:     pastTime,
		HasCookieGetError: true,
		ExpectedErr:       "Unable to renew session",
	},
	{
		Scenario:            "Renew With RememberMe",
		RenewTimeUTC:        pastTime,
		ExpireTimeUTC:       pastTime,
		RememberCookie:      rememberCookie(futureTime, futureTime),
		RenewSessionReturn:  sessionSuccess(futureTime, futureTime),
		GetRememberMeReturn: rememberMe(futureTime, futureTime),
		MethodsCalled:       []string{"GetRememberMe", "RenewSession"},
	},
	{
		Scenario:            "Renew With RememberMe Error",
		RenewTimeUTC:        pastTime,
		ExpireTimeUTC:       pastTime,
		RememberCookie:      rememberCookie(futureTime, futureTime),
		GetRememberMeReturn: rememberMe(futureTime, futureTime),
		RenewSessionReturn:  &SessionReturn{nil, ErrSessionNotFound},
		MethodsCalled:       []string{"GetRememberMe", "RenewSession"},
		ExpectedErr:         "Problem renewing session",
	},
	{
		Scenario:            "Save cookie error",
		RenewTimeUTC:        pastTime,
		ExpireTimeUTC:       pastTime,
		RememberCookie:      rememberCookie(futureTime, futureTime),
		GetRememberMeReturn: rememberMe(futureTime, futureTime),
		RenewSessionReturn:  sessionSuccess(futureTime, futureTime),
		MethodsCalled:       []string{"GetRememberMe", "RenewSession"},
		HasCookiePutError:   true,
		ExpectedErr:         "Error saving session cookie",
	},
}

// NOTE - can't currently get coverage for the error at approx line 147 for the saveSessionCookie error
func TestRenewSession(t *testing.T) {
	for i, test := range renewSessionTests {
		backend := &MockBackend{RenewSessionReturn: test.RenewSessionReturn, GetRememberMeReturn: test.GetRememberMeReturn}
		store := getSessionStore(nil, nil, test.RememberCookie, test.HasCookieGetError, test.HasCookiePutError, backend)
		val, err := store.renewSession("sessionId", "sessionHash", &test.RenewTimeUTC, &test.ExpireTimeUTC)
		methods := store.backend.(*MockBackend).MethodsCalled
		if (err == nil && test.ExpectedErr != "" || err != nil && test.ExpectedErr != err.Error()) ||
			!collectionEqual(test.MethodsCalled, methods) {
			t.Errorf("Scenario[%d] failed: %s\nexpected err:%v\tactual err:%v\nexpected val:%v\tactual val:%v\nexpected methods: %s\tactual methods: %s", i, test.Scenario, test.ExpectedErr, err, test.ExpectedResult, val, test.MethodsCalled, methods)
		}
	}
}

var rememberMeTests = []struct {
	Scenario              string
	HasCookieGetError     bool
	HasCookiePutError     bool
	RememberCookie        *RememberMeCookie
	GetRememberMeReturn   *RememberMeReturn
	RenewRememberMeReturn *RememberMeReturn
	MethodsCalled         []string
	ExpectedResult        *UserLoginRememberMe
	ExpectedErr           string
}{
	{
		Scenario:          "Get RememberMe Cookie err",
		HasCookieGetError: true,
		ExpectedErr:       "RememberMe cookie not found",
	},
	{
		Scenario:            "Renew RememberMe Expired",
		RememberCookie:      rememberCookie(pastTime, pastTime),
		GetRememberMeReturn: rememberMe(pastTime, pastTime),
		ExpectedErr:         "RememberMe cookie has expired",
	},
	{
		Scenario:            "Get RememberMe Error",
		RememberCookie:      rememberCookie(futureTime, futureTime),
		GetRememberMeReturn: &RememberMeReturn{&UserLoginRememberMe{}, ErrRememberMeNotFound},
		MethodsCalled:       []string{"GetRememberMe"},
		ExpectedErr:         "Unable to find matching RememberMe in DB",
	},
	{
		Scenario:            "Get RememberMe Hash Isn't equal",
		RememberCookie:      &RememberMeCookie{"selector", "bogusToken", futureTime, futureTime},
		GetRememberMeReturn: rememberMe(futureTime, futureTime),
		MethodsCalled:       []string{"GetRememberMe"},
		ExpectedErr:         "RememberMe cookie doesn't match backend token",
	},
	{
		Scenario:              "Renew RememberMe Error",
		RememberCookie:        rememberCookie(pastTime, futureTime),
		GetRememberMeReturn:   rememberMe(pastTime, futureTime),
		RenewRememberMeReturn: &RememberMeReturn{&UserLoginRememberMe{}, ErrRememberMeNotFound},
		MethodsCalled:         []string{"GetRememberMe", "RenewRememberMe"},
		ExpectedErr:           "Unable to renew RememberMe",
	},
	{
		Scenario:              "Renew RememberMe Success",
		RememberCookie:        rememberCookie(pastTime, futureTime),
		GetRememberMeReturn:   rememberMe(pastTime, futureTime),
		RenewRememberMeReturn: rememberMe(futureTime, futureTime),
		MethodsCalled:         []string{"GetRememberMe", "RenewRememberMe"},
	},
}

func TestRememberMe(t *testing.T) {
	for i, test := range rememberMeTests {
		backend := &MockBackend{GetRememberMeReturn: test.GetRememberMeReturn, RenewRememberMeReturn: test.RenewRememberMeReturn}
		store := getSessionStore(nil, nil, test.RememberCookie, test.HasCookieGetError, test.HasCookiePutError, backend)
		val, err := store.getRememberMe()
		methods := store.backend.(*MockBackend).MethodsCalled
		if (err == nil && test.ExpectedErr != "" || err != nil && test.ExpectedErr != err.Error()) ||
			!collectionEqual(test.MethodsCalled, methods) {
			t.Errorf("Scenario[%d] failed: %s\nexpected err:%v\tactual err:%v\nexpected val:%v\tactual val:%v\nexpected methods: %s\tactual methods: %s", i, test.Scenario, test.ExpectedErr, err, test.ExpectedResult, val, test.MethodsCalled, methods)
		}
	}
}

var createSessionTests = []struct {
	Scenario              string
	RememberMe            bool
	HasCookieGetError     bool
	HasCookiePutError     bool
	SessionCookie         *SessionCookie
	RememberMeCookie      *RememberMeCookie
	NewLoginSessionReturn *SessionRememberReturn
	MethodsCalled         []string
	ExpectedResult        *UserLoginRememberMe
	ExpectedErr           string
}{
	{
		Scenario:              "New login session error",
		NewLoginSessionReturn: sessionRememberErr(),
		MethodsCalled:         []string{"NewLoginSession"},
		ExpectedErr:           "Unable to create new session",
	},
	{
		Scenario:              "Got session",
		NewLoginSessionReturn: sessionRemember(futureTime, futureTime),
		HasCookieGetError:     true,
		MethodsCalled:         []string{"NewLoginSession"},
	},
	{
		Scenario:              "Valid old session and rememberme cookies.  delete in backend",
		SessionCookie:         sessionCookie(futureTime, futureTime),
		RememberMeCookie:      rememberCookie(futureTime, futureTime),
		NewLoginSessionReturn: sessionRemember(futureTime, futureTime),
		MethodsCalled:         []string{"NewLoginSession", "InvalidateSession", "InvalidateRememberMe"},
	},
	{
		Scenario:              "Set RememberMe",
		RememberMe:            true,
		HasCookieGetError:     true,
		NewLoginSessionReturn: sessionRemember(futureTime, futureTime),
		MethodsCalled:         []string{"NewLoginSession"},
	},
	{
		Scenario:              "Session Cookie save failure",
		HasCookieGetError:     true,
		HasCookiePutError:     true,
		NewLoginSessionReturn: sessionRemember(futureTime, futureTime),
		MethodsCalled:         []string{"NewLoginSession"},
		ExpectedErr:           "Error saving session cookie",
	},
	{
		Scenario:              "RememberMe Cookie save failure",
		RememberMe:            true,
		HasCookieGetError:     true,
		HasCookiePutError:     true,
		NewLoginSessionReturn: sessionRemember(futureTime, futureTime),
		MethodsCalled:         []string{"NewLoginSession"},
		ExpectedErr:           "Unable to save rememberMe cookie",
	},
}

func TestCreateSession(t *testing.T) {
	for i, test := range createSessionTests {
		backend := &MockBackend{NewLoginSessionReturn: test.NewLoginSessionReturn}
		store := getSessionStore(nil, test.SessionCookie, test.RememberMeCookie, test.HasCookieGetError, test.HasCookiePutError, backend)
		val, err := store.CreateSession(1, 1, test.RememberMe)
		methods := store.backend.(*MockBackend).MethodsCalled
		if (err == nil && test.ExpectedErr != "" || err != nil && test.ExpectedErr != err.Error()) ||
			!collectionEqual(test.MethodsCalled, methods) {
			t.Errorf("Scenario[%d] failed: %s\nexpected err:%v\tactual err:%v\nexpected val:%v\tactual val:%v\nexpected methods: %s\tactual methods: %s", i, test.Scenario, test.ExpectedErr, err, test.ExpectedResult, val, test.MethodsCalled, methods)
		}
	}
}

/*************************************************************************************/
type MockSessionStore struct {
	GetSessionReturn    *SessionReturn
	CreateSessionReturn *SessionReturn
}

func (m *MockSessionStore) GetSession() (*UserLoginSession, error) {
	return m.GetSessionReturn.Session, m.GetSessionReturn.Err
}

func (m *MockSessionStore) CreateSession(loginId, userId int, rememberMe bool) (*UserLoginSession, error) {
	return m.CreateSessionReturn.Session, m.CreateSessionReturn.Err
}
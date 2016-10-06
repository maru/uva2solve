// Next problem to solve
// https://github.com/maru/next2solve
//
// Tests for problems.go functionality
//
// To test against the real uHunt API web server:
//   go test next2solve/uhunt -args -real

package uhunt

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"flag"
	"os"
)

//
const (
	// nCPBook3Problems = 1658
	// nUserProblems    = 319
	nUserSubmissions = 729
	problemID        = 1260
	problemNumber    = 10319
	userid           = "46232"
	username         = "chicapi"
)

var (
	testReal bool
)

// Close the test web server
func closeServer(ts *httptest.Server) {
	if ts != nil {
		ts.Close()
	}
}

// HTTP API test server that responds all requests with an invalid response.
func InitAPITestServerInvalid(t *testing.T, apiServer *APIServer, response string) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, response)
	}))
	apiServer.Init(ts.URL)
	return ts
}

// HTTP API test server, real API responses were cached in files.
func InitAPITestServer(t *testing.T, apiServer *APIServer) *httptest.Server {
	// Test against the real uHunt API web server
	if testReal {
		APIUrl := "http://uhunt.felix-halim.net"
		apiServer.Init(APIUrl)
		return nil
	}
	// Create a test API web server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// println(r.RequestURI)
		response, err := ioutil.ReadFile("../test/" + r.RequestURI)
		if err != nil {
			t.Fatalf("Error %v", err)
		}
		fmt.Fprint(w, string(response))
	}))
	apiServer.Init(ts.URL)
	return ts
}

// Test API Server URL
func TestGetUrl(t *testing.T) {
	var apiServer APIServer
	ts := InitAPITestServer(t, &apiServer)
	defer closeServer(ts)

	if ts.URL != apiServer.GetUrl() {
		t.Fatalf("Expected API server URL %s, got %s", ts.URL, apiServer.GetUrl())
	}
}

// Test invalid username
func TestInvalidUsername(t *testing.T) {
	var apiServer APIServer
	ts := InitAPITestServer(t, &apiServer)
	defer closeServer(ts)

	invalidUsername := "not_" + username
	id, err := apiServer.GetUserID(invalidUsername)

	if err != nil {
		t.Fatalf("Error %v", err)
	}
	if id != "0" {
		t.Fatalf("Error expected userid 0")
	}
}

// Test username, invalid response
func TestUsernameInvalidResponse(t *testing.T) {
	var apiServer APIServer
	ts := InitAPITestServerInvalid(t, &apiServer, "")
	defer closeServer(ts)

	id, err := apiServer.GetUserID(username)
	if err != nil {
		t.Fatalf("Error %v", err)
	}
	if id != "" {
		t.Fatalf("Error expected an empty response")
	}
}

// Test valid username
func TestValidUsername(t *testing.T) {
	var apiServer APIServer
	ts := InitAPITestServer(t, &apiServer)
	defer closeServer(ts)

	id, err := apiServer.GetUserID(username)

	if err != nil {
		t.Fatalf("Error %v", err)
	}
	if id != userid {
		t.Fatalf("Error expected userid %s", userid)
	}
}

// Test get problems from CP book
func TestGetCPBookProblems(t *testing.T) {
	var apiServer APIServer
	ts := InitAPITestServer(t, &apiServer)
	defer closeServer(ts)

	problems, err := apiServer.GetProblemListCPbook(3)
	if err != nil {
		t.Fatalf("Error %v", err)
	}
	if problems[0].Title != "Introduction" {
		t.Fatalf("Error %v", "title does not match")
	}
}

// Test get problems from CP book, invalid response
func TestGetCPBookProblemsInvalidResponse(t *testing.T) {
	var apiServer APIServer
	ts := InitAPITestServerInvalid(t, &apiServer, "")
	defer closeServer(ts)

	problems, err := apiServer.GetProblemListCPbook(3)
	if err == nil {
		t.Fatalf("Error %v", "expected end of JSON input")
	}
	if len(problems) != 0 {
		t.Fatalf("Error expected an empty response")
	}

	// empty JSON object
	ts = InitAPITestServerInvalid(t, &apiServer, "{}")
	problems, err = apiServer.GetProblemListCPbook(3)
	if err == nil {
		t.Fatalf("Error %v", "expected json: cannot unmarshal object")
	}
	if len(problems) != 0 {
		t.Fatalf("Error expected an empty response")
	}
}

// Test get problems from CP book, empty number of problems
func TestGetCPBookProblemsEmpty(t *testing.T) {
	var apiServer APIServer
	ts := InitAPITestServerInvalid(t, &apiServer, "[]")
	defer closeServer(ts)

	problems, err := apiServer.GetProblemListCPbook(3)
	if err != nil {
		t.Fatalf("Error %v", err)
	}
	if len(problems) != 0 {
		t.Fatalf("Error expected an empty response")
	}
}

// Test get problem list
func TestGetProblemList(t *testing.T) {
	var apiServer APIServer
	ts := InitAPITestServer(t, &apiServer)
	defer closeServer(ts)

	problems, err := apiServer.GetProblemList()
	if err != nil {
		t.Fatalf("Error %v", err)
	}
	if problems[0].Title != "Introduction" {
		t.Fatalf("Error %v", "title does not match")
	}
}

// Test get user submissions
func TestGetUserSubmissions(t *testing.T) {
	var apiServer APIServer
	ts := InitAPITestServer(t, &apiServer)
	defer closeServer(ts)

	submissions, err := apiServer.GetUserSubmissions(userid)
	if err != nil {
		t.Fatalf("Error %v", err)
	}
	if submissions.Username != username ||
			len(submissions.Submissions) != nUserSubmissions {
		t.Fatalf("Error submissions do not match: read %d expected %d",
			len(submissions.Submissions), nUserSubmissions)
	}
}

// Test get problems from CP book, invalid response
func TestGetUserSubmissionsInvalidResponse(t *testing.T) {
	var apiServer APIServer
	ts := InitAPITestServerInvalid(t, &apiServer, "")
	defer closeServer(ts)

	submissions, err := apiServer.GetUserSubmissions(userid)
	if err == nil {
		t.Fatalf("Error %v", "expected json: cannot unmarshal object")
	}
	if submissions.Username == username ||
			len(submissions.Submissions) != 0 {
			t.Fatalf("Error expected an empty response")
	}

	// empty JSON object
	ts = InitAPITestServerInvalid(t, &apiServer, "[]")

	submissions, err = apiServer.GetUserSubmissions(userid)
	if err == nil {
		t.Fatalf("Error %v", "expected json: cannot unmarshal object")
	}
	if submissions.Username == username ||
			len(submissions.Submissions) != 0 {
			t.Fatalf("Error expected an empty response")
	}
}

// Test get problems from CP book, empty number of problems
func TestGetUserSubmissionsEmpty(t *testing.T) {
	var apiServer APIServer
	ts := InitAPITestServerInvalid(t, &apiServer, "{}")
	defer closeServer(ts)

	submissions, err := apiServer.GetUserSubmissions(userid)
	if err != nil {
		t.Fatalf("Error %v", err)
	}
	if submissions.Username == username ||
			len(submissions.Submissions) != 0 {
			t.Fatalf("Error expected an empty response")
	}
}

// Test get problem details
func TestGetProblemByNum(t *testing.T) {
	var apiServer APIServer
	ts := InitAPITestServer(t, &apiServer)
	defer closeServer(ts)

	problem, err := apiServer.GetProblemByNum(problemNumber)
	if err != nil {
		t.Fatalf("Error %v", err)
	}
	if problem.ProblemID != problemID {
		t.Fatalf("Error %v", "problem does not match")
	}
}

// Test get problem details, invalid response
func TestGetProblemByNumInvalidResponse(t *testing.T) {
	var apiServer APIServer
	ts := InitAPITestServerInvalid(t, &apiServer, "")
	defer closeServer(ts)

	problem, err := apiServer.GetProblemByNum(problemNumber)
	if err == nil {
		t.Fatalf("Error %v", "expected json: cannot unmarshal object")
	}
	if problem.ProblemID == problemID {
			t.Fatalf("Error expected an empty response")
	}

	// empty JSON object
	ts = InitAPITestServerInvalid(t, &apiServer, "[]")

	problem, err = apiServer.GetProblemByNum(problemNumber)
	if err == nil {
		t.Fatalf("Error %v", "expected json: cannot unmarshal object")
	}
	if problem.ProblemID == problemID {
			t.Fatalf("Error expected an empty response")
	}
}

// Test get problem details, empty number of problems
func TestGetProblemByNumEmpty(t *testing.T) {
	var apiServer APIServer
	ts := InitAPITestServerInvalid(t, &apiServer, "{}")
	defer closeServer(ts)

	problem, err := apiServer.GetProblemByNum(problemNumber)
	if err != nil {
		t.Fatalf("Error %v", err)
	}
	if problem.ProblemID == problemID {
			t.Fatalf("Error expected an empty response")
	}
}

// Initialize the test environment
func TestMain(m *testing.M) {
	flag.BoolVar(&testReal, "real", false, "Test with real uHunt API server")
  os.Exit(m.Run())
}

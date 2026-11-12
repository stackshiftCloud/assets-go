package stackshift

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"
)

type computersManifest struct {
	Routes []struct {
		OperationID string `json:"operationId"`
		Method      string `json:"method"`
		Path        string `json:"path"`
		Public      bool   `json:"public"`
	} `json:"routes"`
	Excluded []struct {
		Path string `json:"path"`
	} `json:"excluded"`
}

type routeCall struct {
	Method string
	Path   string
	Auth   string
}

var computerRouteCalls = map[string]func(ctx context.Context, c *ComputersClient) error{
	"registerComputerAgent": func(ctx context.Context, c *ComputersClient) error {
		_, err := RegisterComputerAgent(ctx, c.client.baseURL, nil, ComputerAgentRegistrationInput{Name: "agent"})
		return err
	},
	"getComputerAgentClaim": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.AgentClaim(ctx, "ABCD-1234")
		return err
	},
	"decideComputerAgentClaim": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.DecideAgentClaim(ctx, "ABCD-1234", ComputerAgentClaimDecision{Approve: true})
		return err
	},
	"exchangeComputerAgentClaim": func(ctx context.Context, c *ComputersClient) error {
		_, err := ExchangeComputerAgentClaim(ctx, c.client.baseURL, nil, "claim")
		return err
	},
	"listComputers": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.List(ctx, "p1", ComputerListOptions{})
		return err
	},
	"createComputer": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.Create(ctx, CreateComputerInput{ProjectID: "p1", Name: "n", Template: "coding"}, "")
		return err
	},
	"getComputerApproval": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.Approval(ctx, "a1")
		return err
	},
	"decideComputerApproval": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.DecideApproval(ctx, "a1", true)
		return err
	},
	"getComputerCI": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.CI.Overview(ctx, "p1")
		return err
	},
	"connectComputerCI": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.CI.Connect(ctx, "p1", 7)
		return err
	},
	"disconnectComputerCI": func(ctx context.Context, c *ComputersClient) error {
		return c.CI.Disconnect(ctx, "k1")
	},
	"connectComputerCircleCI": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.CI.ConnectCircleCI(ctx, ComputerCircleCIConnectInput{ProjectID: "p1", ResourceClass: "a/b", Size: "small", APIToken: "a"})
		return err
	},
	"disconnectComputerCircleCI": func(ctx context.Context, c *ComputersClient) error {
		return c.CI.DisconnectCircleCI(ctx, "k1")
	},
	"listComputerCIJobs": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.CI.Jobs(ctx, "p1", 0)
		return err
	},
	"importComputerImage": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.ImportImage(ctx, ImportComputerImageInput{ProjectID: "p1", Image: "node:22", Name: "n"}, "")
		return err
	},
	"listComputerKeys": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.Keys.List(ctx, "p1")
		return err
	},
	"createComputerKey": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.Keys.Create(ctx, CreateComputerKeyInput{ProjectID: "p1", Name: "k"})
		return err
	},
	"revokeComputerKey": func(ctx context.Context, c *ComputersClient) error {
		return c.Keys.Revoke(ctx, "k1")
	},
	"getComputerPricing": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.Pricing(ctx, "p1", "")
		return err
	},
	"getComputerPublicPricing": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.PublicPricing(ctx)
		return err
	},
	"listComputerRegions": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.Regions(ctx, "p1")
		return err
	},
	"getComputerSettings": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.Settings(ctx, "p1")
		return err
	},
	"saveComputerSettings": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.SaveSettings(ctx, "p1", SaveComputerSettingsInput{})
		return err
	},
	"listComputerSnapshots": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.Snapshots.List(ctx, "p1", "", "", 0)
		return err
	},
	"deleteComputerSnapshot": func(ctx context.Context, c *ComputersClient) error {
		return c.Snapshots.Delete(ctx, "s1")
	},
	"createComputerSSHCertificate": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.SSHCertificate(ctx, "c1", "ssh-ed25519 AAAA")
		return err
	},
	"listComputerTemplates": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.Templates(ctx, "p1")
		return err
	},
	"getComputerUsage": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.Usage(ctx, "p1", ComputerUsageOptions{})
		return err
	},
	"deleteComputer": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.Delete(ctx, "c1", "")
		return err
	},
	"getComputer": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.Get(ctx, "c1")
		return err
	},
	"updateComputer": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.Update(ctx, "c1", UpdateComputerInput{})
		return err
	},
	"actComputer": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.Act(ctx, "c1", []ComputerAction{ComputerClick(1, 1)}, "", "")
		return err
	},
	"listComputerActivityLeases": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.ActivityLeases(ctx, "c1")
		return err
	},
	"releaseComputerActivityLease": func(ctx context.Context, c *ComputersClient) error {
		return c.ReleaseActivityLease(ctx, "c1", "build")
	},
	"putComputerActivityLease": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.PutActivityLease(ctx, "c1", "build", 60)
		return err
	},
	"listComputerAgentConversations": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.AgentConversations(ctx, "c1")
		return err
	},
	"streamComputerAgentEvents": func(ctx context.Context, c *ComputersClient) error {
		return c.FollowAgentEvents(ctx, "c1", "v1", 0, 0, func(ComputerAgentEvent) bool { return true })
	},
	"interruptComputerAgent": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.InterruptAgent(ctx, "c1", "v1")
		return err
	},
	"steerComputerAgent": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.SteerAgent(ctx, "c1", "v1", "go left")
		return err
	},
	"promptComputerAgent": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.PromptAgent(ctx, "c1", ComputerAgentPrompt{Prompt: "hi"})
		return err
	},
	"listComputerApprovals": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.Approvals(ctx, "c1", "")
		return err
	},
	"askComputerOwner": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.AskOwner(ctx, "c1", ComputerAskInput{Question: "ok?"})
		return err
	},
	"sendComputerCommandStdin": func(ctx context.Context, c *ComputersClient) error {
		return c.SendStdin(ctx, "c1", "m1", []byte("x"), false)
	},
	"streamComputerEvents": func(ctx context.Context, c *ComputersClient) error {
		return c.Events(ctx, "c1", 0, false, func(ComputerEvent) bool { return true })
	},
	"execComputer": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.Exec(ctx, "c1", ComputerExecInput{Shell: "ls"})
		return err
	},
	"removeComputerFile": func(ctx context.Context, c *ComputersClient) error {
		return c.Remove(ctx, "c1", "a.txt", false)
	},
	"readComputerFile": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.ReadFile(ctx, "c1", "a.txt")
		return err
	},
	"writeComputerFile": func(ctx context.Context, c *ComputersClient) error {
		return c.WriteFile(ctx, "c1", "a.txt", []byte("x"), 0)
	},
	"startComputerDownload": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.StartDownload(ctx, "c1", "a.txt", "")
		return err
	},
	"listComputerFiles": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.ListFiles(ctx, "c1", ".")
		return err
	},
	"makeComputerFolder": func(ctx context.Context, c *ComputersClient) error {
		return c.Mkdir(ctx, "c1", "dir", true)
	},
	"getComputerTransfer": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.Transfer(ctx, "c1", "t1")
		return err
	},
	"startComputerUpload": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.StartUpload(ctx, "c1", ComputerUploadInput{Path: "a.txt", SizeBytes: 1, SHA256: "00"}, "")
		return err
	},
	"completeComputerUpload": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.CompleteUpload(ctx, "c1", "t1", ComputerCompleteUploadInput{UploadID: "u1"})
		return err
	},
	"resumeComputerUpload": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.ResumeUpload(ctx, "c1", "t1", ComputerResumeUploadInput{UploadID: "u1"})
		return err
	},
	"forkComputer": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.Fork(ctx, "c1", ComputerForkInput{Count: 1}, "")
		return err
	},
	"handBackComputer": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.HandBack(ctx, "c1")
		return err
	},
	"listComputerCommandHistory": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.History(ctx, "c1", "", 0)
		return err
	},
	"getComputerCommand": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.Command(ctx, "c1", "")
		return err
	},
	"listComputerJobs": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.Jobs(ctx, "c1")
		return err
	},
	"killComputerJob": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.KillJob(ctx, "c1", "j1", "")
		return err
	},
	"getComputerJobLogs": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.JobLogs(ctx, "c1", "j1", 0)
		return err
	},
	"waitComputerJob": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.WaitJob(ctx, "c1", "j1", 0)
		return err
	},
	"getComputerLimits": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.Limits(ctx, "c1")
		return err
	},
	"listComputerLogs": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.Logs(ctx, "c1", ComputerLogOptions{})
		return err
	},
	"getComputerMetrics": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.Metrics(ctx, "c1", "")
		return err
	},
	"pauseComputer": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.Pause(ctx, "c1", "", "")
		return err
	},
	"pauseComputerAgent": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.PauseAgent(ctx, "c1")
		return err
	},
	"publishComputerPort": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.PublishPort(ctx, "c1", 3000, "")
		return err
	},
	"unpublishComputerPort": func(ctx context.Context, c *ComputersClient) error {
		return c.UnpublishPort(ctx, "c1", 3000)
	},
	"listComputerProcesses": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.Processes(ctx, "c1")
		return err
	},
	"listComputerRecordings": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.Recordings(ctx, "c1")
		return err
	},
	"downloadComputerRecording": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.DownloadRecording(ctx, "c1", "r1")
		return err
	},
	"restoreComputer": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.Restore(ctx, "c1", "s1", "")
		return err
	},
	"resumeComputer": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.Resume(ctx, "c1", "")
		return err
	},
	"resumeComputerAgent": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.ResumeAgent(ctx, "c1")
		return err
	},
	"screenshotComputer": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.Screenshot(ctx, "c1", ComputerScreenshotOptions{})
		return err
	},
	"openComputerSession": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.OpenSession(ctx, "c1", "live")
		return err
	},
	"getComputerSetup": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.Setup(ctx, "c1")
		return err
	},
	"startComputerSetup": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.StartSetup(ctx, "c1", ComputerSetupInput{Script: &ComputerSetupScript{Content: "ls"}}, "")
		return err
	},
	"createComputerSnapshot": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.Snapshot(ctx, "c1", "s", "machine", "")
		return err
	},
	"revokeComputerSSHTokens": func(ctx context.Context, c *ComputersClient) error {
		return c.RevokeSSHTokens(ctx, "c1")
	},
	"createComputerSSHToken": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.CreateSSHToken(ctx, "c1", 0)
		return err
	},
	"listComputerSteps": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.Steps(ctx, "c1", ComputerStepOptions{})
		return err
	},
	"getComputerSystem": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.System(ctx, "c1")
		return err
	},
	"takeOverComputer": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.TakeOver(ctx, "c1")
		return err
	},
	"clearComputerTask": func(ctx context.Context, c *ComputersClient) error {
		return c.ClearTask(ctx, "c1")
	},
	"setComputerTask": func(ctx context.Context, c *ComputersClient) error {
		task := "t"
		_, err := c.SetTask(ctx, "c1", ComputerTaskInput{Task: &task})
		return err
	},
	"listComputerWindows": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.ListWindows(ctx, "c1")
		return err
	},
	"listComputerTeams": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.client.Teams.List(ctx, "p1", TeamListOptions{})
		return err
	},
	"createComputerTeam": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.client.Teams.Create(ctx, CreateTeamInput{ProjectID: "p1", Name: "n", Job: "j", Members: []TeamMemberInput{{Name: "a", Agent: "codex"}}}, "")
		return err
	},
	"getComputerTeam": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.client.Teams.Get(ctx, "t1")
		return err
	},
	"updateComputerTeam": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.client.Teams.Update(ctx, "t1", UpdateTeamInput{})
		return err
	},
	"deleteComputerTeam": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.client.Teams.Delete(ctx, "t1")
		return err
	},
	"pauseComputerTeam": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.client.Teams.Pause(ctx, "t1")
		return err
	},
	"resumeComputerTeam": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.client.Teams.Resume(ctx, "t1")
		return err
	},
	"addComputerTeamMember": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.client.Teams.AddMember(ctx, "t1", TeamMemberInput{Name: "a", Agent: "codex"}, "")
		return err
	},
	"removeComputerTeamMember": func(ctx context.Context, c *ComputersClient) error {
		return c.client.Teams.RemoveMember(ctx, "t1", "m1", false)
	},
	"listComputerTeamMessages": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.client.Teams.ListMessages(ctx, "t1", TeamMessageOptions{})
		return err
	},
	"sendComputerTeamMessage": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.client.Teams.Send(ctx, "t1", SendTeamMessageInput{Text: "hi"})
		return err
	},
	"streamComputerTeamMessages": func(ctx context.Context, c *ComputersClient) error {
		return c.client.Teams.StreamMessages(ctx, "t1", 0, false, func(TeamMessage) bool { return true })
	},
	"listComputerTeamFiles": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.client.Teams.Files(ctx, "t1", "")
		return err
	},
	"getComputerTeamNetwork": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.client.Teams.Network(ctx, "t1")
		return err
	},
	"getComputerTeamActivity": func(ctx context.Context, c *ComputersClient) error {
		_, err := c.client.Teams.Activity(ctx, "t1")
		return err
	},
}

func loadComputersManifest(t *testing.T) computersManifest {
	t.Helper()
	raw, err := os.ReadFile("testdata/computers-routes.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest computersManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	if len(manifest.Routes) == 0 {
		t.Fatal("manifest has no routes")
	}
	return manifest
}

func routePattern(path string) *regexp.Regexp {
	return regexp.MustCompile("^" + regexp.MustCompile(`\\\{[^/]+?\\\}`).ReplaceAllString(regexp.QuoteMeta(path), "[^/]+") + "$")
}

func TestEveryComputersRouteHasASDKMethod(t *testing.T) {
	manifest := loadComputersManifest(t)
	var mu sync.Mutex
	var calls []routeCall
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		calls = append(calls, routeCall{Method: r.Method, Path: strings.TrimPrefix(r.URL.Path, "/api/v1"), Auth: r.Header.Get("Authorization")})
		mu.Unlock()
		if strings.Contains(r.Header.Get("Accept"), "text/event-stream") {
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = w.Write([]byte(": ok\n\n"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet && listRoutes.MatchString(r.URL.Path) {
			_, _ = w.Write([]byte("[]"))
			return
		}
		_, _ = w.Write([]byte("{}"))
	}))
	t.Cleanup(server.Close)
	client, err := New(Options{APIKey: "csk_fixture", BaseURL: server.URL + "/api/v1"})
	if err != nil {
		t.Fatal(err)
	}

	known := map[string]bool{}
	for _, route := range manifest.Routes {
		known[route.OperationID] = true
		call, ok := computerRouteCalls[route.OperationID]
		if !ok {
			t.Errorf("%s %s (%s) has no SDK method in the route table", route.Method, route.Path, route.OperationID)
			continue
		}
		mu.Lock()
		calls = nil
		mu.Unlock()
		if err := call(context.Background(), client.Computers); err != nil {
			t.Errorf("%s: %v", route.OperationID, err)
			continue
		}
		mu.Lock()
		sent := append([]routeCall(nil), calls...)
		mu.Unlock()
		if len(sent) == 0 {
			t.Errorf("%s sent no request", route.OperationID)
			continue
		}
		last := sent[len(sent)-1]
		if last.Method != route.Method || !routePattern(route.Path).MatchString(last.Path) {
			t.Errorf("%s sent %s %s, want %s %s", route.OperationID, last.Method, last.Path, route.Method, route.Path)
		}
		if route.Public && last.Auth != "" {
			t.Errorf("%s is public but sent Authorization", route.OperationID)
		}
	}
	var stale []string
	for id := range computerRouteCalls {
		if !known[id] {
			stale = append(stale, id)
		}
	}
	sort.Strings(stale)
	if len(stale) > 0 {
		t.Errorf("route table has ids missing from the manifest: %v", stale)
	}
}

var listRoutes = regexp.MustCompile(`/computers/(keys|templates)$|/(approvals|recordings|files/list)$`)

var (
	openAPIPathLine   = regexp.MustCompile(`^  (/\S*):\s*$`)
	openAPIMethodLine = regexp.MustCompile(`^    (get|post|put|patch|delete):\s*$`)
)

func TestComputersManifestCoversTheOpenAPIFile(t *testing.T) {
	raw, err := os.ReadFile("../../stackshift_server/openapi/computers.yaml")
	if err != nil {
		t.Skip("openapi/computers.yaml is not checked out next to packages")
	}
	manifest := loadComputersManifest(t)
	listed := map[string]bool{}
	for _, route := range manifest.Routes {
		listed[route.Method+" "+route.Path] = true
	}
	excluded := map[string]bool{}
	for _, entry := range manifest.Excluded {
		excluded[entry.Path] = true
	}
	inPaths, path := false, ""
	for _, line := range strings.Split(string(raw), "\n") {
		if strings.HasPrefix(line, "paths:") {
			inPaths = true
			continue
		}
		if inPaths && len(line) > 0 && line[0] != ' ' {
			break
		}
		if match := openAPIPathLine.FindStringSubmatch(line); match != nil {
			path = match[1]
			continue
		}
		if match := openAPIMethodLine.FindStringSubmatch(line); match != nil && path != "" && !excluded[path] {
			if !listed[strings.ToUpper(match[1])+" "+path] {
				t.Errorf("openapi has %s %s that the manifest does not list", strings.ToUpper(match[1]), path)
			}
		}
	}
}

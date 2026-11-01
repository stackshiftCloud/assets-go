package stackshift

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type DatabaseSleep struct {
	Enabled            bool               `json:"enabled"`
	AfterMinutes       int                `json:"after_minutes"`
	State              string             `json:"state"`
	StateChangedAt     string             `json:"state_changed_at,omitempty"`
	AsleepSince        string             `json:"asleep_since,omitempty"`
	AllowScheduledJobs bool               `json:"allow_scheduled_jobs"`
	LastWakeFailure    *DatabaseWakeEvent `json:"last_wake_failure,omitempty"`
}

type DatabaseSleepInput struct {
	Enabled            bool  `json:"enabled"`
	AfterMinutes       int   `json:"after_minutes,omitempty"`
	AllowScheduledJobs *bool `json:"allow_scheduled_jobs,omitempty"`
}

type DatabaseBranchInfo struct {
	Name             string `json:"name"`
	State            string `json:"state"`
	Mechanism        string `json:"mechanism"`
	ParentDatabaseID string `json:"parent_database_id"`
	ParentName       string `json:"parent_name"`
	RootDatabaseID   string `json:"root_database_id"`
	RootName         string `json:"root_name"`
	Depth            int    `json:"depth"`
}

type Database struct {
	ID     string              `json:"id"`
	Name   string              `json:"name"`
	Type   string              `json:"type"`
	Status string              `json:"status"`
	Sleep  *DatabaseSleep      `json:"sleep,omitempty"`
	Branch *DatabaseBranchInfo `json:"branch,omitempty"`
}

type DatabaseBranchSource struct {
	Kind           string `json:"kind"`
	Time           string `json:"time,omitempty"`
	RestorePointID string `json:"restore_point_id,omitempty"`
	BackupID       string `json:"backup_id,omitempty"`
}

type DatabaseBranchWrites struct {
	Inserted int64 `json:"inserted"`
	Updated  int64 `json:"updated"`
	Deleted  int64 `json:"deleted"`
}

type DatabaseBranchSeed struct {
	DataBytes               int64                 `json:"data_bytes,omitempty"`
	ParentWrites            *DatabaseBranchWrites `json:"parent_writes,omitempty"`
	ScheduledJobsPaused     int                   `json:"scheduled_jobs_paused"`
	SubscriptionsDisabled   int                   `json:"subscriptions_disabled"`
	ReplicationSlotsDropped int                   `json:"replication_slots_dropped"`
	OtherLoginsReset        int                   `json:"other_logins_reset"`
	OutsideServers          []string              `json:"outside_servers,omitempty"`
}

type DatabaseBranchUser struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type DatabaseBranch struct {
	ID               string                   `json:"id"`
	Name             string                   `json:"name"`
	ParentDatabaseID string                   `json:"parent_database_id"`
	RootDatabaseID   string                   `json:"root_database_id"`
	Depth            int                      `json:"depth"`
	Mechanism        string                   `json:"mechanism"`
	Source           DatabaseBranchSource     `json:"source"`
	State            string                   `json:"state"`
	CreatedBy        string                   `json:"created_by"`
	CreatedByUser    *DatabaseBranchUser      `json:"created_by_user,omitempty"`
	ParentName       string                   `json:"parent_name,omitempty"`
	SourceRef        string                   `json:"source_ref,omitempty"`
	ExpiresAt        string                   `json:"expires_at,omitempty"`
	BranchedAt       string                   `json:"branched_at,omitempty"`
	Seed             DatabaseBranchSeed       `json:"seed"`
	FailureCode      string                   `json:"failure_code,omitempty"`
	FailureMessage   string                   `json:"failure_message,omitempty"`
	FailedAt         string                   `json:"failed_at,omitempty"`
	ChangedBytes     *int64                   `json:"changed_bytes"`
	ChangedBytesAt   string                   `json:"changed_bytes_at,omitempty"`
	Cost             *DatabaseBranchCost      `json:"cost,omitempty"`
	CreatedAt        string                   `json:"created_at"`
	UpdatedAt        string                   `json:"updated_at"`
	Database         *Database                `json:"database,omitempty"`
	Operation        *DatabaseBranchOperation `json:"operation,omitempty"`
}

type DatabaseBranchAllowance struct {
	Used         int `json:"used"`
	Limit        int `json:"limit"`
	AccountUsed  int `json:"account_used"`
	AccountLimit int `json:"account_limit"`
	AwakeUsed    int `json:"awake_used"`
	AwakeLimit   int `json:"awake_limit"`
	MaxTTLHours  int `json:"max_ttl_hours"`
}

type DatabaseBranchList struct {
	Branches  []DatabaseBranch        `json:"branches"`
	Allowance DatabaseBranchAllowance `json:"allowance"`
}

type DatabaseBranchWindow struct {
	Earliest string `json:"earliest"`
	Latest   string `json:"latest"`
}

type DatabaseBranchStartingPointChoice struct {
	ID        string `json:"id"`
	At        string `json:"at"`
	Name      string `json:"name,omitempty"`
	Kind      string `json:"kind,omitempty"`
	SizeBytes int64  `json:"size_bytes,omitempty"`
}

type DatabaseBranchStartingPoint struct {
	Available          bool                                `json:"available"`
	Mechanism          string                              `json:"mechanism,omitempty"`
	CopyBytes          int64                               `json:"copy_bytes,omitempty"`
	Window             *DatabaseBranchWindow               `json:"window,omitempty"`
	Choices            []DatabaseBranchStartingPointChoice `json:"choices,omitempty"`
	UnavailableCode    string                              `json:"unavailable_code,omitempty"`
	UnavailableMessage string                              `json:"unavailable_message,omitempty"`
}

type DatabaseBranchStartingPoints struct {
	Now          DatabaseBranchStartingPoint `json:"now"`
	Time         DatabaseBranchStartingPoint `json:"time"`
	RestorePoint DatabaseBranchStartingPoint `json:"restore_point"`
	Backup       DatabaseBranchStartingPoint `json:"backup"`
}

type DatabaseBranchOptions struct {
	StartingPoints           DatabaseBranchStartingPoints `json:"starting_points"`
	Allowance                DatabaseBranchAllowance      `json:"allowance"`
	DefaultTTLHours          int                          `json:"default_ttl_hours"`
	DefaultSleepAfterMinutes int                          `json:"default_sleep_after_minutes"`
	CopyLimitBytes           int64                        `json:"copy_limit_bytes"`
}

type DatabaseScheduledJob struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Schedule string `json:"schedule"`
	Active   bool   `json:"active"`
}

type DatabaseScheduledJobList struct {
	ScheduledJobs []DatabaseScheduledJob `json:"scheduled_jobs"`
	Count         int                    `json:"count,omitempty"`
	ActiveCount   int                    `json:"active_count,omitempty"`
}

type DatabaseBranchResumeScheduledJobsInput struct {
	IDs []int `json:"ids"`
}

type DatabaseBranchTableRows struct {
	Table string `json:"table"`
	Rows  int64  `json:"rows"`
}

type DatabaseBranchTableChange struct {
	Object string `json:"object"`
	Kind   string `json:"kind"`
	Change string `json:"change"`
}

type DatabaseBranchPaused struct {
	ScheduledJobs  []DatabaseScheduledJob `json:"scheduled_jobs"`
	Subscriptions  []string               `json:"subscriptions"`
	OutsideServers []string               `json:"outside_servers"`
}

type DatabaseBranchChanges struct {
	DataBytes              int64                       `json:"data_bytes"`
	ChangedBytes           *int64                      `json:"changed_bytes"`
	ChangedBytesAt         string                      `json:"changed_bytes_at,omitempty"`
	RowsWritten            int64                       `json:"rows_written"`
	RowsWrittenByTable     []DatabaseBranchTableRows   `json:"rows_written_by_table"`
	TableChanges           []DatabaseBranchTableChange `json:"table_changes"`
	TableChangesComplete   bool                        `json:"table_changes_complete"`
	ParentRowsWrittenSince *int64                      `json:"parent_rows_written_since"`
	Paused                 DatabaseBranchPaused        `json:"paused"`
}

type DatabaseWakeEvent struct {
	Path      string `json:"path"`
	Outcome   string `json:"outcome"`
	Reason    string `json:"reason,omitempty"`
	StartedAt string `json:"started_at"`
	ReadyAt   string `json:"ready_at,omitempty"`
	HeldCount int    `json:"held_count"`
}

type DatabaseWakeEventList struct {
	WakeEvents []DatabaseWakeEvent `json:"wake_events"`
}

type DatabaseWakeEventOptions struct {
	Failed bool
	Limit  int
}

type DatabaseBranchStep struct {
	Step string `json:"step"`
	At   string `json:"at"`
}

type DatabaseBranchOperation struct {
	ID               string               `json:"id"`
	BranchDatabaseID string               `json:"branch_database_id"`
	ParentDatabaseID string               `json:"parent_database_id"`
	Kind             string               `json:"kind"`
	Status           string               `json:"status"`
	Phase            string               `json:"phase"`
	Steps            []DatabaseBranchStep `json:"steps,omitempty"`
	ErrorCode        string               `json:"error_code,omitempty"`
	ErrorMessage     string               `json:"error_message,omitempty"`
	CreatedAt        string               `json:"created_at"`
	UpdatedAt        string               `json:"updated_at"`
	CompletedAt      string               `json:"completed_at,omitempty"`
	Branch           *DatabaseBranch      `json:"branch,omitempty"`
}

type DatabaseBranchCreateInput struct {
	Name               string `json:"name"`
	From               any    `json:"from,omitempty"`
	TTL                string `json:"ttl,omitempty"`
	SleepAfterMinutes  int    `json:"sleep_after_minutes,omitempty"`
	CreatedBy          string `json:"created_by,omitempty"`
	SourceRef          string `json:"source_ref,omitempty"`
	ConfirmChargeMinor int    `json:"confirm_charge_minor,omitempty"`
	InstanceClass      string `json:"instance_class,omitempty"`
	IdempotencyKey     string `json:"-"`
}

type DatabaseBranchRetryInput struct {
	ConfirmChargeMinor int    `json:"confirm_charge_minor,omitempty"`
	IdempotencyKey     string `json:"-"`
}

type DatabaseBranchUsdQuote struct {
	MonthlyMinor        int64  `json:"monthly_minor"`
	Hourly              string `json:"hourly"`
	StorageGBMonthMinor int64  `json:"storage_gb_month_minor"`
}

type DatabaseBranchRate struct {
	HourlyMinor         int64                  `json:"hourly_minor"`
	StorageGBMonthMinor int64                  `json:"storage_gb_month_minor"`
	Currency            string                 `json:"currency"`
	USD                 DatabaseBranchUsdQuote `json:"usd"`
}

type DatabaseBranchQuote struct {
	CurrentMonthlyMinor int64              `json:"current_monthly_minor"`
	NewMonthlyMinor     int64              `json:"new_monthly_minor"`
	ChargeMinor         int64              `json:"charge_minor"`
	CreditMinor         int64              `json:"credit_minor"`
	Currency            string             `json:"currency"`
	Billing             string             `json:"billing"`
	Branch              DatabaseBranchRate `json:"branch"`
}

type DatabaseBranchCost struct {
	SoFarMinor          int64                   `json:"so_far_minor"`
	Currency            string                  `json:"currency"`
	HourlyMinor         int64                   `json:"hourly_minor"`
	AwakeSeconds        int64                   `json:"awake_seconds"`
	ComputeMinor        int64                   `json:"compute_minor"`
	ChangedBytes        *int64                  `json:"changed_bytes"`
	StorageMonthlyMinor int64                   `json:"storage_monthly_minor"`
	StorageChargedMinor int64                   `json:"storage_charged_minor"`
	ChargedMinor        int64                   `json:"charged_minor"`
	USD                 *DatabaseBranchUsdQuote `json:"usd,omitempty"`
}

type DatabaseSleepCost struct {
	Billing              string  `json:"billing"`
	Currency             string  `json:"currency"`
	MonthlyMinor         int64   `json:"monthly_minor"`
	AsleepMonthlyMinor   int64   `json:"asleep_monthly_minor"`
	SleepSaves           bool    `json:"sleep_saves"`
	AsleepHoursPerDay    float64 `json:"asleep_hours_per_day"`
	EstimateBasis        string  `json:"estimate_basis"`
	EstimateMonthlyMinor int64   `json:"estimate_monthly_minor"`
}

type DatabaseBranchUsage struct {
	Currency  string `json:"currency"`
	OwedMinor int64  `json:"owed_minor"`
	Stopped   bool   `json:"stopped"`
	StoppedAt string `json:"stopped_at,omitempty"`
	RetireAt  string `json:"retire_at,omitempty"`

	PayerUserID string `json:"payer_user_id,omitempty"`
	PayerName   string `json:"payer_name,omitempty"`
}

type DatabaseBranchUsageCheckout struct {
	Status           string `json:"status"`
	Message          string `json:"message"`
	PaymentURL       string `json:"payment_url,omitempty"`
	PaymentReference string `json:"payment_reference,omitempty"`
	AmountMinor      int64  `json:"amount_minor"`
	WalletMinor      int64  `json:"wallet_minor"`
	Currency         string `json:"currency"`
	PayerUserID      string `json:"payer_user_id,omitempty"`
	PayerName        string `json:"payer_name,omitempty"`
}

type DatabaseBilling struct {
	SleepCost   *DatabaseSleepCost
	BranchUsage *DatabaseBranchUsage
	Raw         map[string]any
}

func (b *DatabaseBilling) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &b.Raw); err != nil {
		return err
	}
	var typed struct {
		SleepCost   *DatabaseSleepCost   `json:"sleep_cost"`
		BranchUsage *DatabaseBranchUsage `json:"branch_usage"`
	}
	if err := json.Unmarshal(data, &typed); err != nil {
		return err
	}
	b.SleepCost = typed.SleepCost
	b.BranchUsage = typed.BranchUsage
	return nil
}

type DatabaseBranchUpdateInput struct {
	Name              string `json:"name,omitempty"`
	TTL               string `json:"ttl,omitempty"`
	SleepAfterMinutes int    `json:"sleep_after_minutes,omitempty"`
}

type DatabaseBranchPromoteInput struct {
	ConfirmName    string `json:"confirm_name"`
	IdempotencyKey string `json:"-"`
}

type DatabaseBranchDeleteOptions struct {
	Cascade        bool
	ConfirmName    string
	IdempotencyKey string
}

type DatabaseBranchWaitOptions struct {
	Timeout  time.Duration
	Interval time.Duration
}

type DatabaseWithBranchOptions struct {
	Create         DatabaseBranchCreateInput
	Direct         bool
	Wait           DatabaseBranchWaitOptions
	CleanupTimeout time.Duration
}

type DatabaseWithBranch struct {
	Branch *DatabaseBranch
	URL    string
}

type DatabaseInstantBranchesMove struct {
	ID           string `json:"id"`
	DatabaseID   string `json:"database_id"`
	Status       string `json:"status"`
	ErrorCode    string `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	CompletedAt  string `json:"completed_at,omitempty"`
}

type DatabaseConnectionEndpoint struct {
	Protocol         string `json:"protocol"`
	Network          string `json:"network"`
	Role             string `json:"role"`
	Primary          bool   `json:"primary"`
	Host             string `json:"host"`
	Port             int    `json:"port"`
	TLSMode          string `json:"tls_mode"`
	TLSServerName    string `json:"tls_server_name,omitempty"`
	ConnectionString string `json:"connection_string"`
	CAFilename       string `json:"ca_filename,omitempty"`
}

type DatabaseCredentials struct {
	Username         string                       `json:"username"`
	Password         string                       `json:"password"`
	DatabaseName     string                       `json:"database_name"`
	Host             string                       `json:"host"`
	Port             int                          `json:"port"`
	TLSMode          string                       `json:"tls_mode"`
	TLSServerName    string                       `json:"tls_server_name,omitempty"`
	CACertificate    string                       `json:"ca_certificate,omitempty"`
	ConnectionString string                       `json:"connection_string"`
	InternalEndpoint *DatabaseConnectionEndpoint  `json:"internal_endpoint,omitempty"`
	PublicEndpoint   *DatabaseConnectionEndpoint  `json:"public_endpoint,omitempty"`
	Endpoints        []DatabaseConnectionEndpoint `json:"endpoints,omitempty"`
}

type DatabaseOperationFailedError struct {
	Operation DatabaseBranchOperation
}

func (e *DatabaseOperationFailedError) Error() string {
	if e.Operation.ErrorMessage != "" {
		return e.Operation.ErrorMessage
	}
	return fmt.Sprintf("stackshift: the %s did not finish", e.Operation.Kind)
}

func (e *DatabaseOperationFailedError) Code() string { return e.Operation.ErrorCode }

type DatabasesClient struct {
	client   *Client
	Branches *DatabaseBranchesClient
}

type DatabaseBranchesClient struct{ client *Client }

func (d *DatabasesClient) Sleep(ctx context.Context, databaseID string, input DatabaseSleepInput) (*DatabaseSleep, error) {
	var out DatabaseSleep
	err := d.client.request(ctx, http.MethodPut, "/databases/"+url.PathEscape(databaseID)+"/sleep", input, &out)
	return &out, err
}

func (d *DatabasesClient) Wake(ctx context.Context, databaseID string, wait bool) (*DatabaseSleep, error) {
	query := url.Values{}
	if wait {
		query.Set("wait", "true")
	}
	var out DatabaseSleep
	err := d.client.request(ctx, http.MethodPost, "/databases/"+url.PathEscape(databaseID)+"/wake"+queryString(query), nil, &out)
	return &out, err
}

func (d *DatabasesClient) ScheduledJobs(ctx context.Context, databaseID string) (*DatabaseScheduledJobList, error) {
	var out DatabaseScheduledJobList
	err := d.client.request(ctx, http.MethodGet, "/databases/"+url.PathEscape(databaseID)+"/scheduled-jobs", nil, &out)
	return &out, err
}

func (d *DatabasesClient) WakeEvents(ctx context.Context, databaseID string, options DatabaseWakeEventOptions) (*DatabaseWakeEventList, error) {
	query := url.Values{}
	if options.Failed {
		query.Set("failed", "true")
	}
	if options.Limit > 0 {
		query.Set("limit", strconv.Itoa(options.Limit))
	}
	var out DatabaseWakeEventList
	err := d.client.request(ctx, http.MethodGet, "/databases/"+url.PathEscape(databaseID)+"/wake-events"+queryString(query), nil, &out)
	return &out, err
}

func (d *DatabasesClient) Billing(ctx context.Context, databaseID string) (*DatabaseBilling, error) {
	var out DatabaseBilling
	err := d.client.request(ctx, http.MethodGet, "/databases/"+url.PathEscape(databaseID)+"/billing", nil, &out)
	return &out, err
}

func (d *DatabasesClient) BranchUsage(ctx context.Context, databaseID string) (*DatabaseBranchUsage, error) {
	var out DatabaseBranchUsage
	err := d.client.request(ctx, http.MethodGet, "/databases/"+url.PathEscape(databaseID)+"/branch-usage", nil, &out)
	return &out, err
}

func (d *DatabasesClient) PayBranchUsage(ctx context.Context, databaseID string) (*DatabaseBranchUsageCheckout, error) {
	var out DatabaseBranchUsageCheckout
	err := d.client.request(ctx, http.MethodPost, "/databases/"+url.PathEscape(databaseID)+"/branch-usage/pay", nil, &out)
	return &out, err
}

func (d *DatabasesClient) MoveToInstantBranches(ctx context.Context, databaseID string) (*DatabaseInstantBranchesMove, error) {
	var out DatabaseInstantBranchesMove
	err := d.client.request(ctx, http.MethodPost, "/databases/"+url.PathEscape(databaseID)+"/move-to-instant-branches", nil, &out)
	return &out, err
}

func (d *DatabasesClient) InstantBranchesMove(ctx context.Context, databaseID string) (*DatabaseInstantBranchesMove, error) {
	var out DatabaseInstantBranchesMove
	err := d.client.request(ctx, http.MethodGet, "/databases/"+url.PathEscape(databaseID)+"/move-to-instant-branches", nil, &out)
	return &out, err
}

func (b *DatabaseBranchesClient) List(ctx context.Context, databaseID string) (*DatabaseBranchList, error) {
	var out DatabaseBranchList
	err := b.client.request(ctx, http.MethodGet, "/databases/"+url.PathEscape(databaseID)+"/branches", nil, &out)
	return &out, err
}

func (b *DatabaseBranchesClient) Create(ctx context.Context, databaseID string, input DatabaseBranchCreateInput) (*DatabaseBranchOperation, error) {
	var out DatabaseBranchOperation
	err := b.client.requestWithHeaders(ctx, http.MethodPost, "/databases/"+url.PathEscape(databaseID)+"/branches", input, idempotencyHeader(input.IdempotencyKey), &out)
	return &out, err
}

func (b *DatabaseBranchesClient) Options(ctx context.Context, databaseID string) (*DatabaseBranchOptions, error) {
	var out DatabaseBranchOptions
	err := b.client.request(ctx, http.MethodGet, "/databases/"+url.PathEscape(databaseID)+"/branches/options", nil, &out)
	return &out, err
}

func (b *DatabaseBranchesClient) Get(ctx context.Context, databaseID, branchID string) (*DatabaseBranch, error) {
	var out DatabaseBranch
	err := b.client.request(ctx, http.MethodGet, "/databases/"+url.PathEscape(databaseID)+"/branches/"+url.PathEscape(branchID), nil, &out)
	return &out, err
}

func (b *DatabaseBranchesClient) Changes(ctx context.Context, databaseID, branchID string) (*DatabaseBranchChanges, error) {
	var out DatabaseBranchChanges
	err := b.client.request(ctx, http.MethodGet, "/databases/"+url.PathEscape(databaseID)+"/branches/"+url.PathEscape(branchID)+"/changes", nil, &out)
	return &out, err
}

func (b *DatabaseBranchesClient) ResumeScheduledJobs(ctx context.Context, databaseID, branchID string, input DatabaseBranchResumeScheduledJobsInput) (*DatabaseScheduledJobList, error) {
	var out DatabaseScheduledJobList
	err := b.client.request(ctx, http.MethodPost, "/databases/"+url.PathEscape(databaseID)+"/branches/"+url.PathEscape(branchID)+"/scheduled-jobs/resume", input, &out)
	return &out, err
}

func (b *DatabaseBranchesClient) Update(ctx context.Context, databaseID, branchID string, input DatabaseBranchUpdateInput) (*DatabaseBranch, error) {
	var out DatabaseBranch
	err := b.client.request(ctx, http.MethodPatch, "/databases/"+url.PathEscape(databaseID)+"/branches/"+url.PathEscape(branchID), input, &out)
	return &out, err
}

func (b *DatabaseBranchesClient) Retry(ctx context.Context, databaseID, branchID string, input DatabaseBranchRetryInput) (*DatabaseBranchOperation, error) {
	var out DatabaseBranchOperation
	err := b.client.requestWithHeaders(ctx, http.MethodPost, "/databases/"+url.PathEscape(databaseID)+"/branches/"+url.PathEscape(branchID)+"/retry", input, idempotencyHeader(defaultString(input.IdempotencyKey, randomHex())), &out)
	return &out, err
}

func (b *DatabaseBranchesClient) Delete(ctx context.Context, databaseID, branchID string, options DatabaseBranchDeleteOptions) (*DatabaseBranchOperation, error) {
	query := url.Values{}
	if options.Cascade {
		if options.ConfirmName == "" {
			return nil, errors.New("stackshift: deleting a branch together with the branches under it needs ConfirmName: pass the name of the branch")
		}
		query.Set("cascade", "true")
		query.Set("confirm", options.ConfirmName)
	}
	var out DatabaseBranchOperation
	err := b.client.requestWithHeaders(ctx, http.MethodDelete, "/databases/"+url.PathEscape(databaseID)+"/branches/"+url.PathEscape(branchID)+queryString(query), nil, idempotencyHeader(defaultString(options.IdempotencyKey, randomHex())), &out)
	return &out, err
}

func (b *DatabaseBranchesClient) Reset(ctx context.Context, databaseID, branchID, idempotencyKey string) (*DatabaseBranchOperation, error) {
	var out DatabaseBranchOperation
	err := b.client.requestWithHeaders(ctx, http.MethodPost, "/databases/"+url.PathEscape(databaseID)+"/branches/"+url.PathEscape(branchID)+"/reset", nil, idempotencyHeader(defaultString(idempotencyKey, randomHex())), &out)
	return &out, err
}

func (b *DatabaseBranchesClient) Promote(ctx context.Context, databaseID, branchID string, input DatabaseBranchPromoteInput) (*DatabaseBranchOperation, error) {
	var out DatabaseBranchOperation
	err := b.client.requestWithHeaders(ctx, http.MethodPost, "/databases/"+url.PathEscape(databaseID)+"/branches/"+url.PathEscape(branchID)+"/promote", input, idempotencyHeader(defaultString(input.IdempotencyKey, randomHex())), &out)
	return &out, err
}

func (b *DatabaseBranchesClient) Operation(ctx context.Context, databaseID, operationID string) (*DatabaseBranchOperation, error) {
	var out DatabaseBranchOperation
	err := b.client.request(ctx, http.MethodGet, "/databases/"+url.PathEscape(databaseID)+"/branches/operations/"+url.PathEscape(operationID), nil, &out)
	return &out, err
}

func (b *DatabaseBranchesClient) WaitForOperation(ctx context.Context, databaseID, operationID string, options DatabaseBranchWaitOptions) (*DatabaseBranchOperation, error) {
	timeout := options.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	interval := options.Interval
	if interval <= 0 {
		interval = time.Second
	}
	deadline := time.Now().Add(timeout)
	for {
		operation, err := b.Operation(ctx, databaseID, operationID)
		if err != nil {
			return nil, err
		}
		if operation.Status == "succeeded" || operation.Status == "failed" {
			return operation, nil
		}
		if !time.Now().Before(deadline) {
			return nil, fmt.Errorf("stackshift: the %s was still %s after %s", operation.Kind, operation.Status, timeout)
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(interval):
		}
	}
}

func (b *DatabaseBranchesClient) URL(ctx context.Context, databaseID, branchID string, direct bool) (string, error) {
	var credentials DatabaseCredentials
	if err := b.client.request(ctx, http.MethodGet, "/databases/"+url.PathEscape(branchID)+"/credentials", nil, &credentials); err != nil {
		return "", err
	}
	role := "pooled"
	if direct {
		role = "direct"
	}
	var match *DatabaseConnectionEndpoint
	for index, endpoint := range credentials.Endpoints {
		if endpoint.Role != role {
			continue
		}
		if match == nil || (endpoint.Network == "public" && match.Network != "public") {
			match = &credentials.Endpoints[index]
		}
	}
	if match != nil {
		return match.ConnectionString, nil
	}
	if direct {
		return credentials.ConnectionString, nil
	}
	return "", errors.New("stackshift: this database has no shared connection string; ask for the direct one")
}

func (b *DatabaseBranchesClient) WithBranch(ctx context.Context, databaseID string, fn func(context.Context, DatabaseWithBranch) error, options ...DatabaseWithBranchOptions) (err error) {
	var settings DatabaseWithBranchOptions
	if len(options) > 0 {
		settings = options[0]
	}
	input := settings.Create
	if input.Name == "" {
		input.Name = "test-" + randomHex()[:8]
	}
	if input.TTL == "" {
		input.TTL = "24h"
	}
	created, err := b.Create(ctx, databaseID, input)
	if err != nil {
		return err
	}
	defer func() {
		cleanup := context.WithoutCancel(ctx)
		cleanupWait := DatabaseBranchWaitOptions{Timeout: settings.CleanupTimeout, Interval: settings.Wait.Interval}
		if cleanupWait.Timeout <= 0 {
			cleanupWait.Timeout = 10 * time.Minute
		}
		_, removeErr := b.WaitForOperation(cleanup, databaseID, created.ID, cleanupWait)
		if removeErr == nil {
			var removal, removed *DatabaseBranchOperation
			if removal, removeErr = b.Delete(cleanup, databaseID, created.BranchDatabaseID, DatabaseBranchDeleteOptions{}); removeErr == nil {
				if removed, removeErr = b.WaitForOperation(cleanup, databaseID, removal.ID, settings.Wait); removeErr == nil && removed.Status == "failed" {
					removeErr = &DatabaseOperationFailedError{Operation: *removed}
				}
			}
		}
		err = errors.Join(err, removeErr)
	}()
	ready, err := b.WaitForOperation(ctx, databaseID, created.ID, settings.Wait)
	if err != nil {
		return err
	}
	if ready.Status == "failed" {
		return &DatabaseOperationFailedError{Operation: *ready}
	}
	branch, err := b.Get(ctx, databaseID, created.BranchDatabaseID)
	if err != nil {
		return err
	}
	connection, err := b.URL(ctx, databaseID, created.BranchDatabaseID, settings.Direct)
	if err != nil {
		return err
	}
	return fn(ctx, DatabaseWithBranch{Branch: branch, URL: connection})
}

func idempotencyHeader(key string) map[string]string {
	if key == "" {
		return nil
	}
	return map[string]string{"Idempotency-Key": key}
}

func randomHex() string {
	raw := make([]byte, 16)
	_, _ = rand.Read(raw)
	return hex.EncodeToString(raw)
}

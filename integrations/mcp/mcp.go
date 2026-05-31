package mcp

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/TrebuchetDynamics/goscrapling"
	"github.com/TrebuchetDynamics/goscrapling/engines/browser"
	"github.com/TrebuchetDynamics/goscrapling/fetchers"
)

const (
	ToolGet               = "get"
	ToolBulkGet           = "bulk_get"
	ToolFetch             = "fetch"
	ToolBulkFetch         = "bulk_fetch"
	ToolStealthyFetch     = "stealthy_fetch"
	ToolBulkStealthyFetch = "bulk_stealthy_fetch"
	ToolScreenshot        = "screenshot"
	ToolOpenSession       = "open_session"
	ToolCloseSession      = "close_session"
	ToolListSessions      = "list_sessions"
)

type ExtractionType string

const (
	ExtractionMarkdown ExtractionType = "markdown"
	ExtractionHTML     ExtractionType = "html"
	ExtractionText     ExtractionType = "text"
)

type SessionType string

const (
	SessionDynamic  SessionType = "dynamic"
	SessionStealthy SessionType = "stealthy"
)

type ContentType string

const (
	ContentText  ContentType = "text"
	ContentImage ContentType = "image"
)

type ResponseModel struct {
	Status  int      `json:"status"`
	Content []string `json:"content"`
	URL     string   `json:"url"`
}

type SessionInfo struct {
	SessionID   string      `json:"session_id"`
	SessionType SessionType `json:"session_type"`
	CreatedAt   string      `json:"created_at"`
	IsAlive     bool        `json:"is_alive"`
}

type SessionCreatedModel struct {
	SessionInfo
	Message string `json:"message"`
}

type SessionClosedModel struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
}

type ContentBlock struct {
	Type     ContentType `json:"type"`
	Text     string      `json:"text,omitempty"`
	MimeType string      `json:"mime_type,omitempty"`
	Data     []byte      `json:"data,omitempty"`
}

type ToolSpec struct {
	Name             string      `json:"name"`
	Description      string      `json:"description,omitempty"`
	InputSchema      InputSchema `json:"input_schema"`
	StructuredOutput bool        `json:"structured_output"`
	ReturnsImage     bool        `json:"returns_image,omitempty"`
}

type InputSchema struct {
	Type       string                    `json:"type"`
	Properties map[string]SchemaProperty `json:"properties"`
	Required   []string                  `json:"required,omitempty"`
}

type SchemaProperty struct {
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
}

type StaticClient interface {
	Get(context.Context, string, fetchers.RequestOptions) (*goscrapling.Response, error)
}

type StaticClientFunc func(context.Context, string, fetchers.RequestOptions) (*goscrapling.Response, error)

func (f StaticClientFunc) Get(ctx context.Context, rawURL string, opts fetchers.RequestOptions) (*goscrapling.Response, error) {
	return f(ctx, rawURL, opts)
}

type BrowserSession interface {
	Fetch(context.Context, string, browser.BrowserOptions) (*goscrapling.Response, error)
	Screenshot(context.Context, string, browser.BrowserOptions) ([]byte, string, error)
	Close(context.Context) error
	Alive() bool
}

type BrowserFactory func(context.Context, SessionType, browser.BrowserOptions, int) (BrowserSession, error)

type ServerOptions struct {
	Static         StaticClient
	BrowserFactory BrowserFactory
}

type Server struct {
	static         StaticClient
	browserFactory BrowserFactory

	mu       sync.Mutex
	sessions map[string]sessionEntry
	counter  int
}

type sessionEntry struct {
	session     BrowserSession
	sessionType SessionType
	createdAt   string
}

func NewServer(opts ServerOptions) *Server {
	static := opts.Static
	if static == nil {
		static = StaticClientFunc(defaultStaticGet)
	}
	factory := opts.BrowserFactory
	if factory == nil {
		factory = func(context.Context, SessionType, browser.BrowserOptions, int) (BrowserSession, error) {
			return nil, fmt.Errorf("mcp browser factory is required for browser tools")
		}
	}
	return &Server{static: static, browserFactory: factory, sessions: make(map[string]sessionEntry)}
}

func defaultStaticGet(ctx context.Context, rawURL string, opts fetchers.RequestOptions) (*goscrapling.Response, error) {
	opts.Context = ctx
	return (fetchers.Fetcher{}).Get(rawURL, opts)
}

func (s *Server) Tools() []ToolSpec {
	return []ToolSpec{
		toolSpec(ToolGet, true, false, []string{"url"}, map[string]string{"url": "string", "css_selector": "string", "main_content_only": "boolean", "extraction_type": "string"}),
		toolSpec(ToolBulkGet, true, false, []string{"urls"}, map[string]string{"urls": "array", "css_selector": "string", "main_content_only": "boolean", "extraction_type": "string"}),
		toolSpec(ToolFetch, true, false, []string{"url"}, map[string]string{"url": "string", "session_id": "string", "css_selector": "string", "main_content_only": "boolean", "extraction_type": "string"}),
		toolSpec(ToolBulkFetch, true, false, []string{"urls"}, map[string]string{"urls": "array", "session_id": "string", "css_selector": "string", "main_content_only": "boolean", "extraction_type": "string"}),
		toolSpec(ToolStealthyFetch, true, false, []string{"url"}, map[string]string{"url": "string", "session_id": "string", "css_selector": "string", "main_content_only": "boolean", "extraction_type": "string"}),
		toolSpec(ToolBulkStealthyFetch, true, false, []string{"urls"}, map[string]string{"urls": "array", "session_id": "string", "css_selector": "string", "main_content_only": "boolean", "extraction_type": "string"}),
		toolSpec(ToolScreenshot, false, true, []string{"url", "session_id"}, map[string]string{"url": "string", "session_id": "string", "image_type": "string", "quality": "integer"}),
		toolSpec(ToolOpenSession, true, false, []string{"session_type"}, map[string]string{"session_type": "string", "session_id": "string", "max_pages": "integer"}),
		toolSpec(ToolCloseSession, true, false, []string{"session_id"}, map[string]string{"session_id": "string"}),
		toolSpec(ToolListSessions, true, false, nil, map[string]string{}),
	}
}

func toolSpec(name string, structured, image bool, required []string, fields map[string]string) ToolSpec {
	properties := make(map[string]SchemaProperty, len(fields))
	for name, typ := range fields {
		properties[name] = SchemaProperty{Type: typ}
	}
	return ToolSpec{Name: name, InputSchema: InputSchema{Type: "object", Properties: properties, Required: append([]string(nil), required...)}, StructuredOutput: structured, ReturnsImage: image}
}

type GetRequest struct {
	URL             string
	ExtractionType  ExtractionType
	CSSSelector     string
	MainContentOnly bool
	Options         fetchers.RequestOptions
}

type BulkGetRequest struct {
	URLs            []string
	ExtractionType  ExtractionType
	CSSSelector     string
	MainContentOnly bool
	Options         fetchers.RequestOptions
}

func (s *Server) Get(ctx context.Context, req GetRequest) (ResponseModel, error) {
	results, err := s.BulkGet(ctx, BulkGetRequest{URLs: []string{req.URL}, ExtractionType: req.ExtractionType, CSSSelector: req.CSSSelector, MainContentOnly: req.MainContentOnly, Options: req.Options})
	if err != nil {
		return ResponseModel{}, err
	}
	if len(results) == 0 {
		return ResponseModel{}, fmt.Errorf("%s: no result", ToolGet)
	}
	return results[0], nil
}

func (s *Server) BulkGet(ctx context.Context, req BulkGetRequest) ([]ResponseModel, error) {
	if len(req.URLs) == 0 {
		return nil, fmt.Errorf("%s: urls is required", ToolBulkGet)
	}
	results := make([]ResponseModel, 0, len(req.URLs))
	for _, rawURL := range req.URLs {
		if strings.TrimSpace(rawURL) == "" {
			return nil, fmt.Errorf("%s: url is required", ToolBulkGet)
		}
		response, err := s.static.Get(ctx, rawURL, req.Options)
		if err != nil {
			return nil, err
		}
		results = append(results, translateResponse(response, req.ExtractionType, req.CSSSelector, req.MainContentOnly))
	}
	return results, nil
}

type FetchRequest struct {
	URL             string
	SessionID       string
	ExtractionType  ExtractionType
	CSSSelector     string
	MainContentOnly bool
	Headless        bool
	Wait            time.Duration
	WaitSelector    string
	NetworkIdle     bool
	Timeout         time.Duration
	Stealth         browser.BrowserStealthOptions
	Options         browser.BrowserOptions
}

type BulkFetchRequest struct {
	URLs            []string
	SessionID       string
	ExtractionType  ExtractionType
	CSSSelector     string
	MainContentOnly bool
	Headless        bool
	Wait            time.Duration
	WaitSelector    string
	NetworkIdle     bool
	Timeout         time.Duration
	Stealth         browser.BrowserStealthOptions
	Options         browser.BrowserOptions
}

func (s *Server) Fetch(ctx context.Context, req FetchRequest) (ResponseModel, error) {
	results, err := s.bulkBrowserFetch(ctx, SessionDynamic, BulkFetchRequest{URLs: []string{req.URL}, SessionID: req.SessionID, ExtractionType: req.ExtractionType, CSSSelector: req.CSSSelector, MainContentOnly: req.MainContentOnly, Headless: req.Headless, Wait: req.Wait, WaitSelector: req.WaitSelector, NetworkIdle: req.NetworkIdle, Timeout: req.Timeout, Options: req.Options})
	if err != nil {
		return ResponseModel{}, err
	}
	return firstResponseModel(ToolFetch, results)
}

func (s *Server) BulkFetch(ctx context.Context, req BulkFetchRequest) ([]ResponseModel, error) {
	return s.bulkBrowserFetch(ctx, SessionDynamic, req)
}

func (s *Server) StealthyFetch(ctx context.Context, req FetchRequest) (ResponseModel, error) {
	results, err := s.bulkBrowserFetch(ctx, SessionStealthy, BulkFetchRequest{URLs: []string{req.URL}, SessionID: req.SessionID, ExtractionType: req.ExtractionType, CSSSelector: req.CSSSelector, MainContentOnly: req.MainContentOnly, Headless: req.Headless, Wait: req.Wait, WaitSelector: req.WaitSelector, NetworkIdle: req.NetworkIdle, Timeout: req.Timeout, Stealth: req.Stealth, Options: req.Options})
	if err != nil {
		return ResponseModel{}, err
	}
	return firstResponseModel(ToolStealthyFetch, results)
}

func (s *Server) BulkStealthyFetch(ctx context.Context, req BulkFetchRequest) ([]ResponseModel, error) {
	return s.bulkBrowserFetch(ctx, SessionStealthy, req)
}

func firstResponseModel(tool string, results []ResponseModel) (ResponseModel, error) {
	if len(results) == 0 {
		return ResponseModel{}, fmt.Errorf("%s: no result", tool)
	}
	return results[0], nil
}

func (s *Server) bulkBrowserFetch(ctx context.Context, sessionType SessionType, req BulkFetchRequest) ([]ResponseModel, error) {
	if len(req.URLs) == 0 {
		return nil, fmt.Errorf("browser fetch: urls is required")
	}
	options := browserOptionsFromFetchRequest(req, sessionType)
	session, closeWhenDone, err := s.browserSessionFor(ctx, sessionType, req.SessionID, options, len(req.URLs))
	if err != nil {
		return nil, err
	}
	if closeWhenDone {
		defer session.Close(ctx)
	}
	results := make([]ResponseModel, 0, len(req.URLs))
	for _, rawURL := range req.URLs {
		response, err := session.Fetch(ctx, rawURL, options)
		if err != nil {
			return nil, err
		}
		results = append(results, translateResponse(response, req.ExtractionType, req.CSSSelector, req.MainContentOnly))
	}
	return results, nil
}

func browserOptionsFromFetchRequest(req BulkFetchRequest, sessionType SessionType) browser.BrowserOptions {
	opts := req.Options
	opts.Headless = req.Headless
	if req.Wait > 0 {
		opts.Wait = req.Wait
	}
	if req.WaitSelector != "" {
		opts.WaitSelector = browser.BrowserWaitSelector{Selector: req.WaitSelector, State: browser.BrowserWaitAttached}
	}
	if req.NetworkIdle {
		opts.NetworkIdle = true
	}
	if req.Timeout > 0 {
		opts.Timeout = req.Timeout
	}
	if sessionType == SessionStealthy {
		opts.Stealth.Enabled = true
		opts.Stealth = mergeMCPStealth(opts.Stealth, req.Stealth)
	}
	return opts
}

func mergeMCPStealth(base, override browser.BrowserStealthOptions) browser.BrowserStealthOptions {
	if override.Enabled {
		base.Enabled = true
	}
	if override.GenerateHeaders {
		base.GenerateHeaders = true
	}
	if override.GoogleReferer {
		base.GoogleReferer = true
	}
	if override.HideCanvas {
		base.HideCanvas = true
	}
	if override.BlockWebRTC {
		base.BlockWebRTC = true
	}
	if override.DisableWebGL {
		base.DisableWebGL = true
	}
	if override.SolveCloudflare {
		base.SolveCloudflare = true
	}
	return base
}

func (s *Server) browserSessionFor(ctx context.Context, sessionType SessionType, sessionID string, options browser.BrowserOptions, maxPages int) (BrowserSession, bool, error) {
	if strings.TrimSpace(sessionID) == "" {
		session, err := s.browserFactory(ctx, sessionType, options, maxPages)
		return session, true, err
	}
	entry, err := s.getSession(sessionID, sessionType)
	if err != nil {
		return nil, false, err
	}
	return entry.session, false, nil
}

type OpenSessionRequest struct {
	SessionType SessionType
	SessionID   string
	Headless    bool
	MaxPages    int
	Options     browser.BrowserOptions
	Stealth     browser.BrowserStealthOptions
}

type CloseSessionRequest struct{ SessionID string }

type ScreenshotRequest struct {
	URL          string
	SessionID    string
	ImageType    string
	FullPage     bool
	Quality      int
	Wait         time.Duration
	WaitSelector string
	NetworkIdle  bool
	Timeout      time.Duration
}

func (s *Server) OpenSession(ctx context.Context, req OpenSessionRequest) (SessionCreatedModel, error) {
	sessionType := req.SessionType
	if sessionType == "" {
		sessionType = SessionDynamic
	}
	if sessionType != SessionDynamic && sessionType != SessionStealthy {
		return SessionCreatedModel{}, fmt.Errorf("unsupported session type %q", sessionType)
	}
	sessionID := strings.TrimSpace(req.SessionID)
	if sessionID == "" {
		sessionID = s.nextSessionID()
	}
	s.mu.Lock()
	if _, exists := s.sessions[sessionID]; exists {
		s.mu.Unlock()
		return SessionCreatedModel{}, fmt.Errorf("session %q already exists", sessionID)
	}
	s.mu.Unlock()

	options := req.Options
	options.Headless = req.Headless
	if sessionType == SessionStealthy {
		options.Stealth.Enabled = true
		options.Stealth = mergeMCPStealth(options.Stealth, req.Stealth)
	}
	maxPages := req.MaxPages
	if maxPages <= 0 {
		maxPages = 5
	}
	session, err := s.browserFactory(ctx, sessionType, options, maxPages)
	if err != nil {
		return SessionCreatedModel{}, err
	}
	createdAt := time.Now().UTC().Format(time.RFC3339Nano)

	s.mu.Lock()
	s.sessions[sessionID] = sessionEntry{session: session, sessionType: sessionType, createdAt: createdAt}
	s.mu.Unlock()
	info := SessionInfo{SessionID: sessionID, SessionType: sessionType, CreatedAt: createdAt, IsAlive: session.Alive()}
	return SessionCreatedModel{SessionInfo: info, Message: fmt.Sprintf("Session %q (%s) created successfully.", sessionID, sessionType)}, nil
}

func (s *Server) nextSessionID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counter++
	return fmt.Sprintf("session-%d", s.counter)
}

func (s *Server) CloseSession(ctx context.Context, req CloseSessionRequest) (SessionClosedModel, error) {
	entry, err := s.removeSession(req.SessionID)
	if err != nil {
		return SessionClosedModel{}, err
	}
	if err := entry.session.Close(ctx); err != nil {
		return SessionClosedModel{}, err
	}
	return SessionClosedModel{SessionID: req.SessionID, Message: fmt.Sprintf("Session %q closed successfully.", req.SessionID)}, nil
}

func (s *Server) ListSessions(context.Context) ([]SessionInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	infos := make([]SessionInfo, 0, len(s.sessions))
	for id, entry := range s.sessions {
		infos = append(infos, SessionInfo{SessionID: id, SessionType: entry.sessionType, CreatedAt: entry.createdAt, IsAlive: entry.session.Alive()})
	}
	return infos, nil
}

func (s *Server) Screenshot(ctx context.Context, req ScreenshotRequest) ([]ContentBlock, error) {
	if req.ImageType == "" {
		req.ImageType = "png"
	}
	if req.ImageType != "png" && req.ImageType != "jpeg" {
		return nil, fmt.Errorf("unsupported image type %q", req.ImageType)
	}
	if req.Quality != 0 && req.ImageType != "jpeg" {
		return nil, fmt.Errorf("quality is only valid for jpeg screenshots")
	}
	entry, err := s.getSession(req.SessionID, "")
	if err != nil {
		return nil, err
	}
	opts := browser.BrowserOptions{Wait: req.Wait, NetworkIdle: req.NetworkIdle, Timeout: req.Timeout, Screenshot: browser.BrowserScreenshotOptions{Enabled: true, FullPage: req.FullPage, Quality: req.Quality}}
	if req.WaitSelector != "" {
		opts.WaitSelector = browser.BrowserWaitSelector{Selector: req.WaitSelector, State: browser.BrowserWaitAttached}
	}
	body, finalURL, err := entry.session.Screenshot(ctx, req.URL, opts)
	if err != nil {
		return nil, err
	}
	return []ContentBlock{{Type: ContentImage, MimeType: "image/" + req.ImageType, Data: append([]byte(nil), body...)}, {Type: ContentText, Text: finalURL}}, nil
}

func (s *Server) getSession(sessionID string, expected SessionType) (sessionEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.sessions[sessionID]
	if !ok {
		return sessionEntry{}, fmt.Errorf("session %q not found", sessionID)
	}
	if !entry.session.Alive() {
		return sessionEntry{}, fmt.Errorf("session %q is no longer alive", sessionID)
	}
	if expected != "" && entry.sessionType != expected {
		return sessionEntry{}, fmt.Errorf("session %q is %q, want %q", sessionID, entry.sessionType, expected)
	}
	return entry, nil
}

func (s *Server) removeSession(sessionID string) (sessionEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.sessions[sessionID]
	if !ok {
		return sessionEntry{}, fmt.Errorf("session %q not found", sessionID)
	}
	delete(s.sessions, sessionID)
	return entry, nil
}

func translateResponse(response *goscrapling.Response, extraction ExtractionType, cssSelector string, mainOnly bool) ResponseModel {
	if response == nil {
		return ResponseModel{}
	}
	content := extractContent(response, extraction, cssSelector, mainOnly)
	return ResponseModel{Status: response.StatusCode(), Content: content, URL: response.URL()}
}

func extractContent(response *goscrapling.Response, extraction ExtractionType, cssSelector string, mainOnly bool) []string {
	selector := strings.TrimSpace(cssSelector)
	if selector == "" && mainOnly {
		selector = "main"
	}
	selection := response.CSS("body")
	if selector != "" {
		selection = response.CSS(selector)
		if selection.Len() == 0 && mainOnly && cssSelector == "" {
			selection = response.CSS("body")
		}
	}
	var content string
	if extraction == ExtractionHTML {
		html, _ := selection.HTML()
		content = html
	} else {
		content = selection.Text()
	}
	content = strings.TrimSpace(content)
	if content == "" {
		return []string{}
	}
	return []string{content}
}

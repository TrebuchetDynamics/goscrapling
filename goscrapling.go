package goscrapling

import (
	"io"

	"github.com/TrebuchetDynamics/goscrapling/core/customtypes"
	"github.com/TrebuchetDynamics/goscrapling/core/storage"
	"github.com/TrebuchetDynamics/goscrapling/engines/browser"
	"github.com/TrebuchetDynamics/goscrapling/engines/toolbelt"
	"github.com/TrebuchetDynamics/goscrapling/fetchers"
	"github.com/TrebuchetDynamics/goscrapling/parser"
)

type ParseOptions = parser.ParseOptions
type SelectorOptions = parser.SelectorOptions
type Document = parser.Document
type Selection = parser.Selection
type Element = parser.Element
type Match = parser.Match
type TextSearchOptions = parser.TextSearchOptions
type SimilarOptions = parser.SimilarOptions
type DiagnosticOptions = parser.DiagnosticOptions
type AdaptiveDiagnostic = parser.AdaptiveDiagnostic
type CandidateDiagnostic = parser.CandidateDiagnostic
type FingerprintFieldDiagnostic = parser.FingerprintFieldDiagnostic

type TextHandler = customtypes.TextHandler
type TextHandlers = customtypes.TextHandlers
type AttributesHandler = customtypes.AttributesHandler

type Key = storage.Key
type Store = storage.Store
type MemoryStore = storage.MemoryStore
type FileStore = storage.FileStore
type SQLiteStore = storage.SQLiteStore
type Fingerprint = storage.Fingerprint
type ScoreComponent = storage.ScoreComponent

type ResponseOptions = toolbelt.ResponseOptions
type RequestMetadata = toolbelt.RequestMetadata
type Response = toolbelt.Response

type BasicAuth = fetchers.BasicAuth
type RequestOptions = fetchers.RequestOptions
type RedirectPolicy = fetchers.RedirectPolicy
type ProxyOptions = fetchers.ProxyOptions
type ProxyRotationStrategy = fetchers.ProxyRotationStrategy
type ProxyRotatorOption = fetchers.ProxyRotatorOption
type ProxyRotator = fetchers.ProxyRotator
type Fetcher = fetchers.Fetcher
type FetcherSessionOptions = fetchers.FetcherSessionOptions
type FetcherSession = fetchers.FetcherSession
type ConcurrentFetcherOptions = fetchers.ConcurrentFetcherOptions
type ConcurrentFetcher = fetchers.ConcurrentFetcher
type ConcurrentRequest = fetchers.ConcurrentRequest
type ConcurrentResult = fetchers.ConcurrentResult
type FetcherErrorKind = fetchers.FetcherErrorKind
type FetcherError = fetchers.FetcherError

type BrowserEngine = browser.BrowserEngine
type BrowserFetcher = browser.BrowserFetcher
type BrowserOptions = browser.BrowserOptions
type BrowserRequest = browser.BrowserRequest
type BrowserResult = browser.BrowserResult
type BrowserWaitState = browser.BrowserWaitState
type BrowserWaitSelector = browser.BrowserWaitSelector
type BrowserActionKind = browser.BrowserActionKind
type BrowserAction = browser.BrowserAction
type ChromedpBrowserOptions = browser.ChromedpBrowserOptions
type ChromedpBrowserEngine = browser.ChromedpBrowserEngine

const (
	DiagnosticBelowThreshold = parser.DiagnosticBelowThreshold
	DiagnosticNoCandidates   = parser.DiagnosticNoCandidates
	DiagnosticMissingStore   = parser.DiagnosticMissingStore

	RedirectPolicySafe = fetchers.RedirectPolicySafe
	RedirectPolicyAll  = fetchers.RedirectPolicyAll
	RedirectPolicyNone = fetchers.RedirectPolicyNone

	FetcherErrorTimeout         = fetchers.FetcherErrorTimeout
	FetcherErrorRetryExhausted  = fetchers.FetcherErrorRetryExhausted
	FetcherErrorRedirect        = fetchers.FetcherErrorRedirect
	FetcherErrorPrivateRedirect = fetchers.FetcherErrorPrivateRedirect
	FetcherErrorProxy           = fetchers.FetcherErrorProxy

	BrowserWaitAttached = browser.BrowserWaitAttached
	BrowserWaitDetached = browser.BrowserWaitDetached
	BrowserWaitVisible  = browser.BrowserWaitVisible
	BrowserWaitHidden   = browser.BrowserWaitHidden

	BrowserActionClick           = browser.BrowserActionClick
	BrowserActionFill            = browser.BrowserActionFill
	BrowserActionWaitForSelector = browser.BrowserActionWaitForSelector
	BrowserActionEvaluate        = browser.BrowserActionEvaluate
)

var (
	ErrMissingStore      = parser.ErrMissingStore
	ErrNilElement        = parser.ErrNilElement
	ErrEmptyIdentifier   = parser.ErrEmptyIdentifier
	ErrInvalidSelector   = parser.ErrInvalidSelector
	ErrInvalidPercentage = parser.ErrInvalidPercentage

	ErrUnsupportedStoreSchema = storage.ErrUnsupportedStoreSchema
	ErrClosedStore            = storage.ErrClosedStore

	ErrRequestOptions                 = fetchers.ErrRequestOptions
	ErrRequestTimeout                 = fetchers.ErrRequestTimeout
	ErrRetryExhausted                 = fetchers.ErrRetryExhausted
	ErrRedirectNotAllowed             = fetchers.ErrRedirectNotAllowed
	ErrPrivateAddressRedirect         = fetchers.ErrPrivateAddressRedirect
	ErrProxyRequest                   = fetchers.ErrProxyRequest
	ErrUnsupportedStaticImpersonation = fetchers.ErrUnsupportedStaticImpersonation
	ErrUnsupportedHTTP3               = fetchers.ErrUnsupportedHTTP3
	ErrMissingBrowserEngine           = browser.ErrMissingBrowserEngine
	NewMemoryStore                    = storage.NewMemoryStore
	NewFileStore                      = storage.NewFileStore
	NewSQLiteStore                    = storage.NewSQLiteStore
	NewResponse                       = toolbelt.NewResponse
	NewFetcherSession                 = fetchers.NewFetcherSession
	NewConcurrentFetcher              = fetchers.NewConcurrentFetcher
	NewProxyRotator                   = fetchers.NewProxyRotator
	WithProxyRotationStrategy         = fetchers.WithProxyRotationStrategy
	CyclicProxyRotation               = fetchers.CyclicProxyRotation
	NewChromedpBrowserEngine          = browser.NewChromedpBrowserEngine
	Bool                              = fetchers.Bool
	NewAttributesHandler              = customtypes.NewAttributesHandler
)

func Parse(r io.Reader, opts ParseOptions) (*Document, error) {
	return parser.Parse(r, opts)
}

func CSSToXPath(selector string) (string, error) {
	return parser.CSSToXPath(selector)
}

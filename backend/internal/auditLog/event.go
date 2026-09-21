package auditlog

const (
	EventUserLogin       = "user.login"
	EventUserLogout      = "user.logout"
	EventUserLoginFailed = "user.login.failed"

	EventDocumentCreate   = "document.create"
	EventDocumentView     = "document.view"
	EventDocumentDownload = "document.download"
	EventDocumentUpdate   = "document.update"
	EventDocumentDelete   = "document.delete"
	EventDocumentShare    = "document.share"

	EventPermissionGrant        = "permission.grant"
	EventPermissionRevoke       = "permission.revoke"
	EventDocumentStatusUpdate   = "document.status.update"
	EventDocumentVersionList    = "document.version.list"
	EventDocumentVersionRestore = "document.version.restore"
	EventDocumentSearch         = "document.search"
	EventDocumentMove           = "document.move"
	EventDocumentRename         = "document.rename"
)

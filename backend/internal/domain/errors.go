package domain

// Error code constants for user-visible errors.
// These are returned to the frontend as error messages,
// and the frontend translates them using i18n keys.
const (
	ErrCommentNotEnabled    = "error.comment.notEnabled"
	ErrPreviewStartFailed   = "error.preview.startFailed"
	ErrCdnNotEnabled        = "error.cdn.notEnabled"
	ErrCdnTokenMissing      = "error.cdn.tokenMissing"
	ErrCdnUploadFailed      = "error.cdn.uploadFailed"
	ErrVercelProjectMissing = "error.deploy.vercelProjectMissing"
	ErrVercelTokenMissing   = "error.deploy.vercelTokenMissing"
	ErrGitTokenMissing      = "error.deploy.gitTokenMissing"
	ErrDeployInProgress     = "error.deploy.inProgress"
	ErrRepoNotConfigured    = "error.deploy.repoNotConfigured"
	ErrNetlifySiteIdMissing = "error.deploy.netlifySiteIdMissing"
	ErrNetlifyTokenMissing  = "error.deploy.netlifyTokenMissing"
	ErrSftpConfigMissing    = "error.deploy.sftpConfigMissing"
	ErrRenderFailed         = "error.render.failed"
)

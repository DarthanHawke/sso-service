package permissions

const (
	// Users
	UserRead   = "users.read"
	UserCreate = "users.create"
	UserUpdate = "users.update"
	UserDelete = "users.delete"

	// Payments
	PaymentCreate = "payments.create"
	PaymentRead   = "payments.read"
	PaymentCancel = "payments.cancel"

	// Admin
	AdminAccess = "admin.access"
)

var All = []string{
	UserRead,
	UserCreate,
	UserUpdate,
	UserDelete,
	PaymentCreate,
	PaymentRead,
	PaymentCancel,
	AdminAccess,
}

package idp

import "context"

type IdentityProviderAPI interface {
	UserManager
	ApplicationManager
}

type UserManager interface {
	CreateUser(ctx context.Context, userInfo *User) (*UserInfo, error)
	GetUser(ctx context.Context, userId string) (*UserInfo, error)
	DeleteUser(ctx context.Context, userId string) error
}

// ApplicationManager defines a contract for all Identity Providers
type ApplicationManager interface {
	GetApplicationInfo(ctx context.Context, applicationId string) (*ApplicationInfo, error)
	CreateApplication(ctx context.Context, app *Application) (*string, error)
	GetApplicationOIDC(ctx context.Context, applicationId string) (*ApplicationOIDCInfo, error)
	DeleteApplication(ctx context.Context, applicationId string) error
}

type User struct {
	FirstName string
	LastName  string
	Email     string
}

type UserInfo struct {
	Id        string
	Email     string
	FirstName string
	LastName  string
}

type Application struct {
	Name        string
	Description string
	TemplateId  string
}

type ApplicationInfo struct {
	Id          string
	Name        string
	Description string
	ClientId    string
}

type ApplicationOIDCInfo struct {
	ClientId     string
	ClientSecret string
}

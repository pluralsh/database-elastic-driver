package driver

import (
	"context"
	"errors"
	"github.com/pluralsh/database-eleastic-driver/pkg/elastic"
	databasespec "github.com/pluralsh/database-interface-api/spec"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"
)

func NewDriver(provisioner string, elasticDB *elastic.Elastic) (*IdentityServer, *ProvisionerServer) {
	return &IdentityServer{
			provisioner: provisioner,
		}, &ProvisionerServer{
			provisioner: provisioner,
			elasticDB:   elasticDB,
		}
}

type ProvisionerServer struct {
	provisioner string
	elasticDB   *elastic.Elastic
}

func (ps *ProvisionerServer) DriverCreateDatabase(ctx context.Context, req *databasespec.DriverCreateDatabaseRequest) (*databasespec.DriverCreateDatabaseResponse, error) {
	if ps.elasticDB == nil {
		klog.Errorf("Database not initialized")
		return &databasespec.DriverCreateDatabaseResponse{}, status.Error(codes.Internal, "Database not initialized")
	}
	databaseName := req.GetName()
	klog.Info("Create Database", "name", databaseName)
	if err := ps.elasticDB.CreateDatabase(ctx, databaseName); err != nil {
		if status.Code(err) != codes.AlreadyExists {
			klog.Errorf("Failed to create database %v", err)
			return &databasespec.DriverCreateDatabaseResponse{}, err
		}
	}

	return &databasespec.DriverCreateDatabaseResponse{
		DatabaseId: databaseName,
	}, nil
}

func (ps *ProvisionerServer) DriverDeleteDatabase(ctx context.Context, req *databasespec.DriverDeleteDatabaseRequest) (*databasespec.DriverDeleteDatabaseResponse, error) {
	if ps.elasticDB == nil {
		klog.Errorf("Database not initialized")
		return &databasespec.DriverDeleteDatabaseResponse{}, status.Error(codes.Internal, "Database not initialized")
	}
	klog.Info("Delete Database", "name", req.DatabaseId)
	if err := ps.elasticDB.DeleteDatabase(ctx, req.DatabaseId); err != nil {
		klog.Errorf("Failed to delete database %v", err)
		return &databasespec.DriverDeleteDatabaseResponse{}, err
	}

	return &databasespec.DriverDeleteDatabaseResponse{}, nil
}

// This call grants access to an account. The account_name in the request shall be used as a unique identifier to create credentials.
// The account_id returned in the response will be used as the unique identifier for deleting this access when calling DriverRevokeDatabaseAccess.
func (ps *ProvisionerServer) DriverGrantDatabaseAccess(_ context.Context, req *databasespec.DriverGrantDatabaseAccessRequest) (*databasespec.DriverGrantDatabaseAccessResponse, error) {
	klog.Info("Grant access", "name", req.DatabaseId)
	resp := &databasespec.DriverGrantDatabaseAccessResponse{
		AccountId:   req.DatabaseId,
		Credentials: map[string]*databasespec.CredentialDetails{},
	}
	resp.Credentials["cred"] = &databasespec.CredentialDetails{Secrets: map[string]string{
		"DB_USER":     ps.elasticDB.User,
		"DB_PASSWORD": ps.elasticDB.Password,
		"DB_ADDRESS":  ps.elasticDB.Address,
		"INDEX_NAME":  req.DatabaseId,
	}}

	return resp, nil
}

// This call revokes all access to a particular database from a principal.
func (ps *ProvisionerServer) DriverRevokeDatabaseAccess(context.Context, *databasespec.DriverRevokeDatabaseAccessRequest) (*databasespec.DriverRevokeDatabaseAccessResponse, error) {
	return &databasespec.DriverRevokeDatabaseAccessResponse{}, nil
}

type IdentityServer struct {
	provisioner string
}

func (id *IdentityServer) DriverGetInfo(context.Context, *databasespec.DriverGetInfoRequest) (*databasespec.DriverGetInfoResponse, error) {
	if id.provisioner == "" {
		klog.Error(errors.New("provisioner name cannot be empty"), "Invalid argument")
		return nil, status.Error(codes.InvalidArgument, "ProvisionerName is empty")
	}

	return &databasespec.DriverGetInfoResponse{
		Name: id.provisioner,
	}, nil
}

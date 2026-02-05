package aws_only

import (
	"context"
	"testing"
	"tests/utils"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func TestCreateUser(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)

	const (
		region   = "sa-east-1"
		endpoint = "http://localhost:4566"
		table    = "table1"
		userPK   = "user#123"
	)

	db, err := utils.NewDynamoDB(ctx, region, endpoint)
	if err != nil {
		t.Fatalf("NewDynamoDB(region=%s, endpoint=%s) failed: %v", region, endpoint, err)
	}

	t.Run("dynamodb is reachable and required tables exist", func(t *testing.T) {
		t.Helper()

		if err := db.AssertTablesExist(ctx, []string{table}, 10*time.Second); err != nil {
			existing, listErr := db.ListTables(ctx)
			if listErr != nil {
				t.Fatalf(
					"expected table %q to exist, but AssertTablesExist failed: %v (also failed to ListTables: %v)",
					table, err, listErr,
				)
			}
			t.Fatalf(
				"expected table %q to exist, but AssertTablesExist failed: %v. existing tables: %v",
				table, err, existing,
			)
		}

		t.Logf("table %q exists in LocalStack DynamoDB (region=%s, endpoint=%s)", table, region, endpoint)
	})

	t.Run("dynamodb getting userPK successfull", func(t *testing.T) {
		// 1) sanity: tabela existe
		assertTablesExist(t, ctx, db, []string{table}, 10*time.Second)

		// 2) busca item
		key := map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: userPK},
		}

		item, err := db.GetItem(ctx, table, key)
		if err != nil {
			t.Fatalf("GetItem(table=%s, pk=%s) failed: %v", table, userPK, err)
		}
		if len(item) == 0 {
			t.Fatalf("expected item to exist in table %q for pk=%q, but item was empty", table, userPK)
		}

		// 3) asserts de propriedade
		requireAttrStringEquals(t, item, "status", "ACTIVE")
		requireAttrExists(t, item, "createdAt")          // só verificar que existe
		requireAttrStringNonEmpty(t, item, "email")      // existe e não é vazio
		requireAttrBoolEquals(t, item, "isActive", true) // exemplo boolean
	})
}

// ---------- helpers de testes (sem libs externas) ----------

func assertTablesExist(t *testing.T, ctx context.Context, db *utils.DynamoClient, tables []string, timeout time.Duration) {
	t.Helper()

	if err := db.AssertTablesExist(ctx, tables, timeout); err != nil {
		existing, listErr := db.ListTables(ctx)
		if listErr != nil {
			t.Fatalf("AssertTablesExist(tables=%v) failed: %v (also ListTables failed: %v)", tables, err, listErr)
		}
		t.Fatalf("AssertTablesExist(tables=%v) failed: %v (existing=%v)", tables, err, existing)
	}
}

func requireAttrExists(t *testing.T, item map[string]types.AttributeValue, attr string) {
	t.Helper()

	if _, ok := item[attr]; !ok {
		t.Fatalf("expected attribute %q to exist, but it was missing. item=%v", attr, item)
	}
}

func requireAttrStringEquals(t *testing.T, item map[string]types.AttributeValue, attr, want string) {
	t.Helper()

	av, ok := item[attr]
	if !ok {
		t.Fatalf("expected attribute %q to exist, but it was missing. item=%v", attr, item)
	}

	s, ok := av.(*types.AttributeValueMemberS)
	if !ok {
		t.Fatalf("expected attribute %q to be string (S), got %T", attr, av)
	}

	if s.Value != want {
		t.Fatalf("attribute %q mismatch: want=%q got=%q", attr, want, s.Value)
	}
}

func requireAttrStringNonEmpty(t *testing.T, item map[string]types.AttributeValue, attr string) {
	t.Helper()

	av, ok := item[attr]
	if !ok {
		t.Fatalf("expected attribute %q to exist, but it was missing. item=%v", attr, item)
	}

	s, ok := av.(*types.AttributeValueMemberS)
	if !ok {
		t.Fatalf("expected attribute %q to be string (S), got %T", attr, av)
	}

	if s.Value == "" {
		t.Fatalf("expected attribute %q to be non-empty, but it was empty", attr)
	}
}

func requireAttrBoolEquals(t *testing.T, item map[string]types.AttributeValue, attr string, want bool) {
	t.Helper()

	av, ok := item[attr]
	if !ok {
		t.Fatalf("expected attribute %q to exist, but it was missing. item=%v", attr, item)
	}

	b, ok := av.(*types.AttributeValueMemberBOOL)
	if !ok {
		t.Fatalf("expected attribute %q to be BOOL, got %T", attr, av)
	}

	if b.Value != want {
		t.Fatalf("attribute %q mismatch: want=%v got=%v", attr, want, b.Value)
	}
}

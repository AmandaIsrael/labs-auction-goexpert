package auction

import (
	"context"
	"testing"
	"time"

	"fullcycle-auction_go/internal/entity/auction_entity"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestAuctionRepository_CreateAuction_ShouldCloseAutomatically_WhenDurationExpires(t *testing.T) {
	ctx := context.Background()

	mongoContainer, err := mongodb.Run(ctx, "mongo:7")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = mongoContainer.Terminate(ctx)
	})

	uri, err := mongoContainer.ConnectionString(ctx)
	require.NoError(t, err)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = client.Disconnect(ctx)
	})

	database := client.Database("auctions_test")

	t.Setenv("AUCTION_DURATION", "2s")
	repository := NewAuctionRepository(database)

	auction, errEntity := auction_entity.CreateAuction(
		"Notebook Dell",
		"informatica",
		"Notebook Dell i7 16GB seminovo",
		auction_entity.Used,
	)
	require.Nil(t, errEntity)

	errCreate := repository.CreateAuction(ctx, auction)
	require.Nil(t, errCreate)

	created, errFind := repository.FindAuctionById(ctx, auction.Id)
	require.Nil(t, errFind)
	require.Equal(t, auction_entity.Active, created.Status,
		"auction should start as Active right after creation")

	require.Eventually(t, func() bool {
		current, errCurrent := repository.FindAuctionById(ctx, auction.Id)
		if errCurrent != nil {
			return false
		}
		return current.Status == auction_entity.Completed
	}, 6*time.Second, 200*time.Millisecond,
		"auction status should automatically become Completed after AUCTION_DURATION")
}

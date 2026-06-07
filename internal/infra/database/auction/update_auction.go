package auction

import (
	"context"
	"fmt"
	"fullcycle-auction_go/configuration/logger"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/internal_error"

	"go.mongodb.org/mongo-driver/bson"
)

func (ar *AuctionRepository) UpdateAuctionStatus(
	ctx context.Context,
	auctionId string,
	status auction_entity.AuctionStatus) *internal_error.InternalError {
	filter := bson.M{"_id": auctionId}
	update := bson.M{"$set": bson.M{"status": status}}

	if _, err := ar.Collection.UpdateOne(ctx, filter, update); err != nil {
		logger.Error(
			fmt.Sprintf("Error trying to update auction status with id = %s", auctionId), err)
		return internal_error.NewInternalServerError("Error trying to update auction status")
	}

	return nil
}

package handler

import (
	"context"
	"log"
	"time"

	"hackitbackend/app/models"
	"hackitbackend/pkg/utils"
	"hackitbackend/platform/database"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ctx = context.Background()

func Index(c *fiber.Ctx) error {
	return utils.ResponseMsg(c, 200, "Api is running", nil)
}

func UpdateUsers(c *fiber.Ctx) error {
	_id := c.Params("id")
	users := new(models.Users)
	db, err := database.Connect()

	if err != nil {
		log.Fatal(err.Error())
	}

	if err := c.BodyParser(users); err != nil {
		return utils.ResponseMsg(c, 400, "error", err.Error())
	} else {
		if docID, err := primitive.ObjectIDFromHex(_id); err != nil {
			return utils.ResponseMsg(c, 400, "error", err)
		} else {
			q := bson.M{"_id": docID}

			u := bson.D{
				{Key: "$set", Value: bson.D{
					{Key: "username", Value: users.Username},
					{Key: "email", Value: users.Email},
					{Key: "updatedAt", Value: time.Now()},
				},
				}}

			o := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

			db.Collection("users").FindOneAndUpdate(ctx, q, u, o).Decode(&users)
			return utils.ResponseMsg(c, 400, "Update Data Succesfully", users)
		}
	}
}

func GetByIdUsers(c *fiber.Ctx) error {
	_id := c.Params("id")
	var ctx = context.Background()
	db, err := database.Connect()

	if err != nil {
		log.Fatal(err.Error())
	}

	if docID, err := primitive.ObjectIDFromHex(_id); err != nil {
		return utils.ResponseMsg(c, 400, "Get Data unsuccesfully", err.Error())
	} else {
		q := bson.M{"_id": docID}
		users := models.Users{}
		result := db.Collection("users").FindOne(ctx, q)
		result.Decode(&users)
		if result.Err() != nil {
			return utils.ResponseMsg(c, 200, "Get Data unsuccesfully", result.Err().Error())
		} else {
			return utils.ResponseMsg(c, 200, "Get Data Succesfully", users)
		}
	}

}

func GetUsers(c *fiber.Ctx) error {

	db, err := database.Connect()

	if err != nil {
		log.Fatal(err.Error())
	}

	csr, err := db.Collection("users").Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err.Error())
	}

	defer csr.Close(ctx)

	result := make([]models.Users, 0)
	for csr.Next(ctx) {
		var row models.Users
		err := csr.Decode(&row)
		if err != nil {
			log.Fatal(err.Error())
		}

		result = append(result, row)
	}

	return utils.ResponseMsg(c, 200, "Get Data Succesfully", result)
}

func CreateUsers(c *fiber.Ctx) error {
	users := new(models.Users)
	var ctx = context.Background()
	db, err := database.Connect()

	if err != nil {
		log.Fatal(err.Error())
	}

	users.CreatedAt = time.Now()
	users.UpdatedAt = time.Now()

	if err := c.BodyParser(users); err != nil {
		return utils.ResponseMsg(c, 400, err.Error(), nil)
	} else {
		if r, err := db.Collection("users").InsertOne(ctx, users); err != nil {
			return utils.ResponseMsg(c, 500, "Inserted data unsuccesfully", err.Error())
		} else {
			return utils.ResponseMsg(c, 200, "Inserted data succesfully", r)
		}
	}
}

func DeleteUsers(c *fiber.Ctx) error {
	_id := c.Params("id")
	var ctx = context.Background()
	db, err := database.Connect()

	if err != nil {
		log.Fatal(err.Error())
	}

	if docID, err := primitive.ObjectIDFromHex(_id); err != nil {
		return utils.ResponseMsg(c, 400, "Sucess", nil)
	} else {
		q := bson.M{"_id": docID}
		r := db.Collection("users").FindOneAndDelete(ctx, q)
		if r.Err() != nil {
			return utils.ResponseMsg(c, 400, "error", r)
		} else {
			return utils.ResponseMsg(c, 200, "Sucess", r)
		}
	}
}

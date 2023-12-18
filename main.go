package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var collection *mongo.Collection

type Car struct {
	ID     string `json:"car_id" bson:"_id,omitempty"`
	Brand  string `json:"brand" bson:"brand"`
	Model  string `json:"model" bson:"model"`
	Status string `json:"status" bson:"status"`
}

func init() {
	// Replace the following with your MongoDB connection URI
	clientOptions := options.Client().ApplyURI("mongodb+srv://notexist123:notexist123@cluster0.v1k4aao.mongodb.net/garage")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	collection = client.Database("garage").Collection("cars")
}

func main() {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.POST("/cars", addCar)
	r.GET("/cars", getCars)
	r.PUT("/cars/:car_id", updateCarStatus)
	r.DELETE("/cars/:car_id", deleteCar)
	r.DELETE("/cars", deleteAllCars)

	if err := r.Run(":5000"); err != nil {
		log.Fatal(err)
	}
}

func addCar(c *gin.Context) {
	var car Car
	if err := c.ShouldBindJSON(&car); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	if car.Brand == "" || car.Model == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Brand and Model are required"})
		return
	}

	existingCar := findCarByBrandAndModel(car.Brand, car.Model)
	if existingCar.ID != "" {
		c.JSON(http.StatusConflict, gin.H{"error": "Car already exists"})
		return
	}

	car.Status = "In Garage" // Initial status
	result, err := collection.InsertOne(context.Background(), car)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Car added successfully", "car_id": result.InsertedID})
}

func getCars(c *gin.Context) {
	var cars []Car
	cur, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	defer cur.Close(context.Background())

	for cur.Next(context.Background()) {
		var car Car
		err := cur.Decode(&car)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}
		cars = append(cars, car)
	}

	c.JSON(http.StatusOK, gin.H{"cars": cars})
}

func updateCarStatus(c *gin.Context) {
	carID := c.Param("car_id")
	var data map[string]string
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	newStatus, exists := data["status"]
	if !exists || (newStatus != "In Garage" && newStatus != "Under Repair" && newStatus != "Completed") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
		return
	}

	result, err := collection.UpdateOne(context.Background(), bson.M{"_id": carID}, bson.M{"$set": bson.M{"status": newStatus}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Car not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Car status updated successfully"})
}

func deleteCar(c *gin.Context) {
	carID := c.Param("car_id")
	result, err := collection.DeleteOne(context.Background(), bson.M{"_id": carID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Car not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Car deleted successfully"})
}

func deleteAllCars(c *gin.Context) {
	result, err := collection.DeleteMany(context.Background(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%d cars deleted successfully", result.DeletedCount)})
}

func findCarByBrandAndModel(brand, model string) Car {
	var car Car
	err := collection.FindOne(context.Background(), bson.M{"brand": brand, "model": model}).Decode(&car)
	if err != nil {
		return Car{}
	}
	return car
}

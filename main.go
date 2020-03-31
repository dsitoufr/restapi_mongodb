package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// product model
type Product struct {
	ID          string    `json:"id,omitempty" bson:"id,omitempty"`
	ProductName string    `json:"prodname,omitempty" bson:"prodname,omitempty"`
	UnitPrice   string    `json:"prodprice,omitempty" bson:"prodprice,omitempty"`
	Supplier    *Supplier `json:"supplier,omitempty" bson:"supplier,omitempty"`
}

// supplier model
type Supplier struct {
	Name    string `json:"name,omitempty" bson:"name,omitempty"`
	Country string `json:"country,omitempty" bson:"country,omitempty"`
}

//list of products
var (
	products []Product
	client   *mongo.Client
)

// Get one single product
func getOneProduct(response http.ResponseWriter, request *http.Request) {
	var product Product

	response.Header().Set("Content-Type", "application/json")
	params := mux.Vars(request)
	//id := params["id"]

	collection := client.Database("golangdb").Collection("products")
	err := collection.FindOne(context.TODO(), Product{ID: params["id"]}).Decode(&product)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Found a single product: %+v\n", product)
	json.NewEncoder(response).Encode(product)
}

// Get all products
func getAllProducts(response http.ResponseWriter, request *http.Request) {
	var products []Product

	findOptions := options.Find()
	findOptions.SetLimit(2)

	response.Header().Set("Content-Type", "application/json")

	collection := client.Database("golangdb").Collection("products")
	cur, err := collection.Find(context.TODO(), bson.D{{}}, findOptions)
	if err != nil {
		log.Fatal(err)
	}

	defer cur.Close(context.TODO())
	for cur.Next(context.TODO()) {
		var product Product
		err := cur.Decode(&product)
		if err != nil {
			log.Fatal(err)
		}
		products = append(products, product)
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	json.NewEncoder(response).Encode(products)
	log.Printf("find multiply documents: %+v\n", products)
}

//create new product
func newProduct(response http.ResponseWriter, request *http.Request) {
	var product Product

	response.Header().Set("Content-Type", "application/json")
	_ = json.NewDecoder(request.Body).Decode(&product)
	product.ID = strconv.Itoa(rand.Intn(100000000))

	collection := client.Database("golangdb").Collection("products")

	results, err := collection.InsertOne(context.TODO(), product)
	if err != nil {
		log.Printf("failed to insert a doc: %v", err)
		return
	}

	log.Println("Inserted a document: ", results.InsertedID)
	json.NewEncoder(response).Encode(product)

}

// Update Product
func updateProduct(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	var product Product

	params := mux.Vars(request)
	id := params["id"]
	prodprice := params["prodprice"]

	_ = json.NewDecoder(request.Body).Decode(&product)

	filter := bson.D{{"id", id}}
	update := bson.D{{"$set", bson.D{{"prodprice", prodprice}}}}

	collection := client.Database("golangdb").Collection("products")
	results, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Matched documents %v  updated documents %v.\n", results.MatchedCount, results.ModifiedCount)

	err = collection.FindOne(context.TODO(), Product{ID: id}).Decode(&product)
	if err != nil {
		log.Fatal(err)
	}

	json.NewEncoder(response).Encode(product)

}

// Delete Product
func deleteProduct(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")

	params := mux.Vars(request)
	id := params["id"]
	delete := bson.D{{"id", id}}

	collection := client.Database("golangdb").Collection("products")
    
	results, err := collection.DeleteOne(context.TODO(),delete)
     if err != nil {
		 log.Fatal(err)
	 }

	 log.Printf("deleted %v documents\n", results.DeletedCount)
	//json.NewEncoder(response).Encode(products)
}

// Main function
func main() {
	var err error

	// Set mongodb client options
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	// Connect to mongodb
	client, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Check connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to MongoDB")

	// Init router
	router := mux.NewRouter()

	// Route handles & endpoints
	router.HandleFunc("/api/product", newProduct).Methods("POST")
	router.HandleFunc("/api/product/{id}", deleteProduct).Methods("DELETE")
	router.HandleFunc("/api/product/{id}", getOneProduct).Methods("GET")
	router.HandleFunc("/api/products", getAllProducts).Methods("GET")
	router.HandleFunc("/api/product/{id}/{prodprice}", updateProduct).Methods("PUT")

	// Start server
	log.Println("Starting server...")
	log.Fatal(http.ListenAndServe(":8000", router))

}

//{
//	"prodname":"table",
//	"prodprice":"1500",
//	"supplier":{"name":"Harry","country":"USA"}
//}

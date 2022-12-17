# Inventory golang rest service

### Requirements to run the app in docker
* Docker should be installed.
* Make should be installed.

## How to start

Run below command that will build and start the go application inside docker container.

``
make run
``

Run the below command to stop the application from running.

``
make stop
``

## Attached postman collection for east testing.
## Attached swagger specification yaml for better understanding.

## Application Usage
### Using this inventory you can achieve following operations
1. Add an item to the inventory
   1. validates request
   2. post validation a unique id will be assigned to the item and stored in the inventory list/map
   3. using list to persist the order and for better pagination, using map for quick retrivals
2. Get all items from inventory
   1. takes 3 optional query parameter for pagination, page and size of integer data type.
3. Get single item using item id.
   1. validates id
   2. checks id is present in the system
   3. return item
4. Delete item by id.
   1. validates id
   2. checks id is present in the system else throw not found error.
   3. delete from the list and map
5. Update a item by id.
   1. validates id
   2. validates item in the request body
   3. checks if id is exists else throw error.
   4. update the item in both array and map
6. Download csv
   1. downloads all the items in a csv
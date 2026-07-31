# 👕 FitStore-engine (Admin-backend)

A lightweight administrative backend engine built with **Go (Golang)**, the **Echo web framework**, and **GORM ORM** using a **PostgreSQL** database.

---

## 🛠️ Clone the Project

Open your terminal workspace and run the following commands to clone the repository and navigate into it:

```bash
git clone https://github.com/Abhi071998/Tshirt-admin-backend.git
cd tshirt-go
```

## 🐘  Database Setup (TablePlus)
Before running the backend, you must create the database container inside PostgreSQL:

Open TablePlus and connect to your local PostgreSQL server instance.

Press Ctrl + G (or Cmd + G on Mac) to open the database panel.

Click "New..." or "+", type the database name exactly as tshirt_store, and click Save/Create.
## .env setup
create a .env file inside a root directory

Copy the env variables which are there in .env.example and add your credentials

## 🚀 Step 4: Run the Application
While still inside the apps/admin-backend directory, run the following commands to download dependencies and launch the engine:
```bash
#  Clean up and download required packages
go mod tidy
# Start the server
go run cmd/api/main.go
```

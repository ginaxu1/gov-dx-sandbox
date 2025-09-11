import json
import strawberry
from typing import Optional
import uvicorn
from strawberry.fastapi import GraphQLRouter
from fastapi import FastAPI
from dotenv import load_dotenv

# SQLAlchemy imports for table creation
from mock_data import Father, Informant, Mother, mock_data
from pydantic import BaseModel
from datetime import date
from contextlib import asynccontextmanager

from mock_data import PersonData

load_dotenv()


from strawberry.types import Info

@strawberry.federation.type
class Query:
    @strawberry.field(description="Get person information by NIC")
    def health_check(self) -> str:
        return "Healthy"

    @strawberry.field(description="Get person information by NIC")
    def get_person_info(self, nic: strawberry.ID) -> Optional[PersonData]:
        # Get Data From Mock Data
        for data in mock_data['birth']:
            if data.nic == str(nic):
                return data
        return None

schema = strawberry.federation.Schema(query=Query, types=[PersonData, Informant, Father, Mother])

@asynccontextmanager
async def lifespan(app: FastAPI):
    openapi_schema = app.openapi()
    with open("openapi.json", "w") as f:
        json.dump(openapi_schema, f, sort_keys=False)
    print("✅ OpenAPI schema written to openapi.json")
    # Setup code
    yield
    # Teardown code

# Create FastAPI app
app = FastAPI(
    title="Mock RGD GraphQL API",
    description="Mock Registrar General's Department GraphQL subgraph providing person address and profession data",
    version="1.0.0",
    openapi_url="/openapi.json",
    docs_url="/docs",
    lifespan=lifespan
)

# Pydantic schema for request validation
from typing import Optional
class PersonCreate(BaseModel):
    full_name: str
    other_names: Optional[str] = None
    birth_date: Optional[date] = None
    birth_place: Optional[str] = None
    email: str
    nic: str
    address: Optional[str] = None
    profession: Optional[str] = None

# Add GraphQL router
graphql_app = GraphQLRouter(schema)
app.include_router(graphql_app, prefix="/graphql")

# Health check endpoint
@app.get("/health")
async def health_check():
    return {"status": "healthy", "service": "mock-rgd"}

# Root endpoint with service info
@app.get("/")
async def root():
    return {
        "service": "Mock RGD GraphQL API",
        "description": "Provides person address and profession data by NIC",
        "endpoints": {
            "graphql": "/graphql",
            "health": "/health"
        }
    }


# read port from environment variable
import os
port = int(os.getenv("PORT", 8080))

if __name__ == "__main__":
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=8080,
        reload=True,
        log_level="info"
    )

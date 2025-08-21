import json
import strawberry
from typing import Optional
import uvicorn
from strawberry.fastapi import GraphQLRouter
from fastapi import FastAPI, Depends, HTTPException
from dotenv import load_dotenv

# SQLAlchemy imports for table creation
from database import engine, get_db
from models import SQLAlchemyPersonInfo
from sqlalchemy.orm import Session
from pydantic import BaseModel
from datetime import date
from contextlib import asynccontextmanager

load_dotenv()

# Strawberry GraphQL type
@strawberry.type
class PersonInfo:
    nic: str = strawberry.field(description="National Identity Card number")
    address: str = strawberry.field(description="Person's address")
    profession: str = strawberry.field(description="Person's profession")


from strawberry.types import Info

@strawberry.type
class Query:
    @strawberry.field(description="Get person information by NIC")
    def person(self, info: Info[None, None], nic: strawberry.ID) -> Optional[PersonInfo]:
        db: Session = info.context["db"]
        person = db.query(SQLAlchemyPersonInfo).filter_by(nic=nic).first()
        if person:
            return PersonInfo(
                nic=person.nic,
                address=person.address,
                profession=person.profession
            )
        return None

    @strawberry.field(description="Get all available person records")
    def all_persons(self, info: Info[None, None]) -> list[PersonInfo]:
        db: Session = info.context["db"]
        people = db.query(SQLAlchemyPersonInfo).all()
        return [PersonInfo(nic=p.nic, address=p.address, profession=p.profession) for p in people]



# Strawberry context for DB session
def get_context_dependency(db: Session = Depends(get_db)):
    return {"db": db}

schema = strawberry.federation.Schema(query=Query)

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

# Create all tables in the database
# Create all tables in the database
from database import Base
Base.metadata.create_all(bind=engine)


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


# POST endpoint to create a new Person
@app.post("/person", response_model=dict, tags=["Person"])
def create_person(person: PersonCreate, db: Session = Depends(get_db)):
    db_person = SQLAlchemyPersonInfo(
        full_name=person.full_name,
        other_names=person.other_names,
        birth_date=person.birth_date,
        birth_place=person.birth_place,
        email=person.email,
        nic=person.nic,
        address=person.address,
        profession=person.profession
    )
    db.add(db_person)
    try:
        db.commit()
        db.refresh(db_person)
    except Exception as e:
        db.rollback()
        raise HTTPException(status_code=400, detail=str(e))
    return {"id": db_person.id, "nic": db_person.nic}

# Add GraphQL router
graphql_app = GraphQLRouter(schema, context_getter=get_context_dependency)
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

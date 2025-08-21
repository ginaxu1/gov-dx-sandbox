import os
from dotenv import load_dotenv
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker, declarative_base

load_dotenv()

DB_USER = os.getenv('CHOREO_MOCK_DMT_DB_CONNECTION_USERNAME', 'your_username')
DB_PASSWORD = os.getenv('CHOREO_MOCK_DMT_DB_CONNECTION_PASSWORD', 'your_password')
DB_HOST = os.getenv('CHOREO_MOCK_DMT_DB_CONNECTION_HOSTNAME', 'localhost')
DB_PORT = os.getenv('CHOREO_MOCK_DMT_DB_CONNECTION_PORT', '5432')
DB_NAME = os.getenv('CHOREO_MOCK_DMT_DB_CONNECTION_DATABASENAME', 'your_database')

DATABASE_URL = f'postgresql://{DB_USER}:{DB_PASSWORD}@{DB_HOST}:{DB_PORT}/{DB_NAME}'

engine = create_engine(DATABASE_URL, echo=True)
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)
Base = declarative_base()

def get_db():
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()


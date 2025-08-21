from sqlalchemy import Column, Integer, String, Date
from database import Base

class SQLAlchemyPersonInfo(Base):
    __tablename__ = 'person_info'
    id = Column(Integer, primary_key=True, index=True)
    full_name = Column(String, nullable=False)
    other_names = Column(String, nullable=True)
    birth_date = Column(Date)
    birth_place = Column(String, nullable=True)
    email = Column(String, unique=True, index=True, nullable=False)
    nic = Column(String, unique=True, index=True, nullable=False)
    address = Column(String)
    profession = Column(String)

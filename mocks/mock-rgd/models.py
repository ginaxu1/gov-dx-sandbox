from sqlalchemy import Column, Integer, String, Date, Boolean
from sqlalchemy.orm import relationship
from sqlalchemy import ForeignKey
from database import Base

class SQLAlchemyFatherInfo(Base):
    __tablename__ = 'father_info'
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String, nullable=False)
    nic = Column(String, unique=True, index=True, nullable=False)
    birth_date = Column(Date, nullable=True)
    birth_place = Column(String, nullable=True)
    race = Column(String, nullable=True)

    children = relationship("SQLAlchemyPersonInfo", back_populates="father")

class SQLAlchemyMotherInfo(Base):
    __tablename__ = 'mother_info'
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String, nullable=False)
    birth_date = Column(Date, nullable=True)
    birth_place = Column(String, nullable=True)
    race = Column(String, nullable=True)
    nic = Column(String, unique=True, index=True, nullable=False)
    age_at_birth = Column(Integer, nullable=True)

    children = relationship("SQLAlchemyPersonInfo", back_populates="mother")

class SQLAlchemyPersonInfo(Base):
    __tablename__ = 'person_info'
    id = Column(Integer, primary_key=True, index=True)
    brNo = Column(String)
    district = Column(String)
    division = Column(String)
    birth_date = Column(Date)
    birth_place = Column(String, nullable=True)
    name = Column(String, nullable=False)
    sex = Column(String, nullable=True)
    other_names = Column(String, nullable=True)
    email = Column(String, unique=True, index=True, nullable=False)
    nic = Column(String, unique=True, index=True, nullable=False)
    address = Column(String)
    profession = Column(String)
    are_parents_married = Column(Boolean, nullable=True)
    is_grandfather_born_in_sri_lanka = Column(Boolean, nullable=True)

    father_id = Column(Integer, ForeignKey('father_info.id'), nullable=True)
    mother_id = Column(Integer, ForeignKey('mother_info.id'), nullable=True)

    father = relationship("SQLAlchemyFatherInfo", back_populates="children")
    mother = relationship("SQLAlchemyMotherInfo", back_populates="children")
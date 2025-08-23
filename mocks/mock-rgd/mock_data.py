from dataclasses import dataclass
from datetime import date
import strawberry
from typing import Optional
from strawberry.federation.schema_directives import External

@strawberry.type
class Informant:
    signature: str
    full_name: str
    residence: str
    relationship_to_baby: str
    nic: str

@strawberry.type
class Father:
    name: str
    nic: str
    birth_date: date
    birth_place: str
    race: str

@strawberry.type
class Mother:
    name: str
    nic: str
    birth_date: date
    birth_place: str
    race: str
    age_at_birth: int

@dataclass
class PersonInfo:
    id: int
    brNo: str
    district: str
    division: str
    birth_date: date
    birth_place: str
    name: str
    sex: str
    nic: str
    are_parents_married: bool
    is_grandfather_born_in_sri_lanka: bool
    father: Father
    mother: Mother
    date_of_registration: date
    registrar_signature: str
    informant: Informant

@strawberry.federation.type(keys=["nic"])
class PersonData:
    nic: strawberry.ID
    id: int
    brNo: str
    district: str
    division: str
    birth_date: date
    birth_place: str
    name: str
    sex: str
    are_parents_married: bool
    is_grandfather_born_in_sri_lanka: bool
    father: Father
    mother: Mother
    date_of_registration: date
    registrar_signature: str
    informant: Informant

        # Reference resolver (called by federation gateway)
    @classmethod
    def resolve_reference(cls, nic: strawberry.ID) -> Optional["PersonData"]:
        # Federation passes { nic: "..." } here

        for person in mock_data["birth"]:
            if person.nic == nic:
                return person

        return None


mock_data = {
    "birth": [
        PersonData(
            id=1,
            brNo="BR2025001",
            district="Colombo",
            division="Colombo North",
            birth_date=date(2020, 5, 12),
            birth_place="Colombo General Hospital",
            name="Aarav Perera",
            sex="Male",
            nic=strawberry.ID("199512345678"),
            are_parents_married=True,
            is_grandfather_born_in_sri_lanka=True,
            father=Father(
                name="Sunil Perera",
                nic="710123456V",
                birth_date=date(1985, 8, 21),
                birth_place="Colombo",
                race="Sinhala"
            ),
            mother=Mother(
                name="Kamala Perera",
                nic="790987654V",
                birth_date=date(1987, 2, 11),
                birth_place="Galle",
                race="Sinhala",
                age_at_birth=33
            ),
            date_of_registration=date(2020, 5, 15),
            registrar_signature="R. Silva",
            informant=Informant(
                signature="Sunil Perera",
                full_name="Sunil Perera",
                residence="12 Galle Rd, Colombo",
                relationship_to_baby="Father",
                nic="710123456V"
            )
        ),
        PersonData(
            id=2,
            brNo="BR2025002",
            district="Galle",
            division="Galle Urban",
            birth_date=date(2021, 1, 20),
            birth_place="Galle General Hospital",
            name="Nisha Fernando",
            sex="Female",
            nic=strawberry.ID("210120678V"),
            are_parents_married=True,
            is_grandfather_born_in_sri_lanka=False,
            father=Father(
                name="Dinesh Fernando",
                nic="680123456V",
                birth_date=date(1982, 4, 10),
                birth_place="Matara",
                race="Sinhala"
            ),
            mother=Mother(
                name="Shalini Fernando",
                nic="750987654V",
                birth_date=date(1985, 12, 5),
                birth_place="Galle",
                race="Sinhala",
                age_at_birth=36
            ),
            date_of_registration=date(2021, 1, 22),
            registrar_signature="M. Jayawardena",
            informant=Informant(
                signature="Dinesh Fernando",
                full_name="Dinesh Fernando",
                residence="45 Main St, Galle",
                relationship_to_baby="Father",
                nic="680123456V"
            )
        ),
        PersonData(
            id=3,
            brNo="BR2025003",
            district="Kandy",
            division="Kandy Central",
            birth_date=date(2019, 8, 9),
            birth_place="Kandy Teaching Hospital",
            name="Rohan Jayasuriya",
            sex="Male",
            nic=strawberry.ID("190809234V"),
            are_parents_married=False,
            is_grandfather_born_in_sri_lanka=True,
            father=Father(
                name="Chaminda Jayasuriya",
                nic="730567890V",
                birth_date=date(1980, 6, 2),
                birth_place="Kandy",
                race="Sinhala"
            ),
            mother=Mother(
                name="Anjali Jayasuriya",
                nic="760987123V",
                birth_date=date(1982, 11, 20),
                birth_place="Nuwara Eliya",
                race="Tamil",
                age_at_birth=37
            ),
            date_of_registration=date(2019, 8, 12),
            registrar_signature="P. De Silva",
            informant=Informant(
                signature="Anjali Jayasuriya",
                full_name="Anjali Jayasuriya",
                residence="23 Temple Rd, Kandy",
                relationship_to_baby="Mother",
                nic="760987123V"
            )
        ),
        PersonData(
            id=4,
            brNo="BR2025004",
            district="Galle",
            division="Galle South",
            birth_date=date(2020, 1, 15),
            birth_place="Galle General Hospital",
            name="Ayesha Mohamed",
            sex="Female",
            nic=strawberry.ID("200101234V"),
            are_parents_married=True,
            is_grandfather_born_in_sri_lanka=True,
            father=Father(
                name="Mohamed Ali",
                nic="680123456V",
                birth_date=date(1985, 5, 10),
                birth_place="Galle",
                race="Sri Lankan Moor"
            ),
            mother=Mother(
                name="Fatima Ali",
                nic="750987654V",
                birth_date=date(1988, 8, 20),
                birth_place="Galle",
                race="Sri Lankan Moor",
                age_at_birth=32
            ),
            date_of_registration=date(2020, 1, 20),
            registrar_signature="https://example.com/signatures/680123456V",
            informant=Informant(
                signature="https://example.com/signatures/680123456V",
                full_name="Mohamed Ali",
                residence="45 Main St, Galle",
                relationship_to_baby="Father",
                nic="680123456V"
            )
        )
    ]
}

# translate data type from PersonData to PersonData

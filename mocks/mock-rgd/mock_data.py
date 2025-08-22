from datetime import date
import strawberry

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

@strawberry.type
class PersonData:
    id: int
    brNo: str
    district: str
    division: str
    birth_date: date
    birth_place: str
    name: str
    sex: str
    other_names: str
    email: str
    nic: str
    profession: str
    are_parents_married: bool
    is_grandfather_born_in_sri_lanka: bool
    father: Father
    mother: Mother
    date_of_registration: date
    registrar_signature: str
    informant: Informant

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
            other_names="Arav",
            email="aarav.perera@example.com",
            nic="200512345V",
            profession="N/A",
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
            other_names="Nishi",
            email="nisha.fernando@example.com",
            nic="210120678V",
            profession="N/A",
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
            other_names="",
            email="rohan.jayasuriya@example.com",
            nic="190809234V",
            profession="N/A",
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
            brNo="8150",
            district="Puttalam",
            division="Chilaw",
            birth_date=date(2000, 7, 18),
            birth_place="Chilaw General Hospital",
            name="Mohamed Mushraf",
            sex="Male",
            other_names="Mushi",
            email="mushrafmim@gmail.com",
            nic="200020000500",
            profession="N/A",
            are_parents_married=True,
            is_grandfather_born_in_sri_lanka=True,
            father=Father(
                name="Seyyadhu Hussain Mohamed Ismail",
                nic="196230701022",
                birth_date=date(1962, 11, 2),
                birth_place="Chilaw",
                race="Sri Lankan Moor"
            ),
            mother=Mother(
                name="Nasrin Thajudeen",
                nic="750987654V",
                birth_date=date(1965, 4, 4),
                birth_place="Sawarana",
                race="Sri Lankan Moor",
                age_at_birth=35
            ),
            date_of_registration=date(2000, 9, 1),
            registrar_signature="https://example.com/signatures/196230701022",
            informant=Informant(
                signature="https://example.com/signatures/196230701022",
                full_name="S. H. M. Ismail",
                residence="Wayalthottam, Sawarana, Chilaw.",
                relationship_to_baby="Father",
                nic="196230701022"
            )
        )
    ]
}
# Project Initialization
1. Install PostgreSQL
```bash
brew install postgresql
```
2. setup your username and password for postgresql.
3. create a `Config.toml` file inside `mocks/mock-dmt/` folder and place the values inside `Config.toml`
```toml
[mock_dmt.store]
host = ""
port = 5432
user = ""
password = ""
database = "mock-dmt"
```
4. Load the `mocks/mock-dmt/modules/store/script.sql` file.
5. Run the server
```bash
bal run
```
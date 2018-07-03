# **MySQL Database Updater** [![Build Status](https://travis-ci.org/catdevman/mysql-database-updater.svg?branch=master)](https://travis-ci.org/catdevman/mysql-database-updater)

## Usage
|   Flag   | Value Type  |                          Description                       |     Default    |
|----------|-------------|------------------------------------------------------------|----------------|
| dbPrefix |    string   | Choose a prefix for the databases that this will loop over |   db_          |
| env      |    string   | Choose environment from environments file                  |  local         |
| envFile  |    string   | Choose path for environments file                          |environments.csv|
| sqlFile  |    string   | Choose path for sql file                                   |  updates.sql   |

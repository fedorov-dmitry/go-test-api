CREATE SCHEMA IF NOT EXISTS app;

CREATE TABLE IF NOT EXISTS app.currency_rates (
  date      date          NOT NULL,
  base      varchar(5)    NOT NULL,
  currency  varchar(5)    NOT NULL,
  rate      numeric       NOT NULL,
  PRIMARY KEY (date, base, currency)
);



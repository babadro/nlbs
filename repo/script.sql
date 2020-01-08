create table transactions
(
    amount         numeric(7, 2) not null,
    cancelled      boolean       not null,
    transaction_id bigint        not null
        constraint transactions_pk
            primary key
);

alter table transactions
    owner to postgres;

create table users
(
    balance numeric(7, 2) not null,
    id      serial        not null
        constraint users_pk
            primary key
);

alter table users
    owner to postgres;

create or replace function transactions_cancel() returns void
    language plpgsql
as
$$
DECLARE
    total      integer       := (SELECT COUNT(*)
                                 FROM transactions);
    last_count integer       := 10;
    skip_count integer       := total - last_count;
    amountSum  numeric(7, 2) := 0;
BEGIN
    IF skip_count < 0 THEN
        skip_count = 0;
    END IF;
    CREATE TEMP TABLE last_transactions
    (
        transaction_id bigint,
        amount         numeric(7, 2),
        cancelled      bool
    ) on commit drop;

    INSERT INTO last_transactions
    SELECT transaction_id, amount, cancelled
    FROM transactions OFFSET skip_count
    LIMIT last_count;
    SELECT COALESCE(SUM(last_transactions.amount), 0)
    INTO amountSum
    FROM last_transactions
    WHERE last_transactions.cancelled = false;
    UPDATE users
    SET balance = balance - amountSum
    WHERE id = 1;
    UPDATE transactions
    SET cancelled = true
    WHERE transaction_id IN (SELECT transaction_id FROM last_transactions WHERE cancelled = false);
END;
$$;

alter function transactions_cancel() owner to postgres;

create or replace function balance_update(tr_amount numeric, tr_id bigint) returns void
    language plpgsql
as
$$
DECLARE
    old_balance numeric(7, 2);
    new_balance numeric(7, 2);
BEGIN
    LOCK TABLE users IN ACCESS EXCLUSIVE MODE;
    SELECT balance INTO old_balance FROM users WHERE id = 1;
    new_balance = old_balance + tr_amount;
    if new_balance < 0 THEN
        new_balance = 0;
        tr_amount = -old_balance;
    end if;
    UPDATE users SET balance = new_balance WHERE id = 1;
    INSERT INTO transactions (amount, transaction_id, cancelled) VALUES (tr_amount, tr_id, false);
END;
$$;

alter function balance_update(numeric, bigint) owner to postgres;



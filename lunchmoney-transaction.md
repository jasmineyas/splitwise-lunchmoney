Insert Transactions
Use this endpoint to insert one or more transactions at once. The maximum number of transactions per requests is 500.

HTTP Request
POST https://dev.lunchmoney.app/v1/transactions

Body Parameters

Parameter Type Required Default Description
transactions array true - List of transactions to insert (see below)
apply_rules boolean false false If true, will apply account’s existing rules to the inserted transactions. Defaults to false.
skip_duplicates boolean false false If true, the system will automatically dedupe based on transaction date, payee and amount. Note that deduping by external_id will occur regardless of this flag.
check_for_recurring boolean false false If true, will check new transactions for occurrences of new monthly expenses. Defaults to false.
debit_as_negative boolean false false If true, will assume negative amount values denote expenses and positive amount values denote credits. Defaults to false.
skip_balance_update boolean false true If true, will skip updating balance if an asset_id is present for any of the transactions.
Transaction Object to Insert

Key Type Required Description
date string true Must be in ISO 8601 format (YYYY-MM-DD).
amount number/string true Numeric value of amount. i.e. $4.25 should be denoted as 4.25.
category_id number false Unique identifier for associated category_id. Category must be associated with the same account and must not be a category group.
payee string false Max 140 characters
currency string false Three-letter lowercase currency code in ISO 4217 format. The code sent must exist in our database. Defaults to user account's primary currency.
asset_id number false Unique identifier for associated asset (manually-managed account). Asset must be associated with the same account. When set plaid_account_id may not be set.
plaid_account_id number false Unique identifier for associated plaid account. This must have the "Allow Modifications to Transactions" option set. When set asset_id may not be set.
recurring_id number false Unique identifier for associated recurring expense. Recurring expense must be associated with the same account.
notes string false Max 350 characters
status string false Must be either cleared or uncleared. (Note: special statuses for recurring items have been deprecated.) Defaults to uncleared.
external_id string false User-defined external ID for transaction. Max 75 characters. External IDs must be unique within the same asset_id.
tags Array of numbers and/or strings false Passing in a number will attempt to match by ID. If no matching tag ID is found, an error will be thrown. Passing in a string will attempt to match by string. If no matching tag name is found, a new tag will be created.

Example 200 Response
Upon success, IDs of inserted transactions will be returned in an array.
{
"ids": [54, 55, 56, 57]
}

Example 404 Response
An array of errors will be returned denoting reason why parameters were deemed invalid.
{
"error": [
"Transaction 0 is missing date.",
"Transaction 0 is missing amount.",
"Transaction 1 status must be either cleared or uncleared: null"
]
}

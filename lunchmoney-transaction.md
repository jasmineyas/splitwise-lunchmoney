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

```json
{
            "id": 2271300563,
            "date": "2025-01-02",
            "amount": "75.0000",
            "currency": "cad",
            "to_base": 75,
            "payee": "Find & Save",
            "category_id": null,
            "category_name": null,
            "category_group_id": null,
            "category_group_name": null,
            "is_income": false,
            "exclude_from_budget": false,
            "exclude_from_totals": false,
            "created_at": "2025-04-16T15:54:15.367Z",
            "updated_at": "2025-04-16T15:54:15.367Z",
            "status": "uncleared",
            "is_pending": false,
            "notes": null,
            "original_name": "to Find & Save",
            "recurring_id": null,
            "recurring_payee": null,
            "recurring_description": null,
            "recurring_cadence": null,
            "recurring_granularity": null,
            "recurring_quantity": null,
            "recurring_type": null,
            "recurring_amount": null,
            "recurring_currency": null,
            "parent_id": null,
            "has_children": false,
            "group_id": null,
            "is_group": false,
            "asset_id": null,
            "asset_institution_name": null,
            "asset_name": null,
            "asset_display_name": null,
            "asset_status": null,
            "plaid_account_id": 162269,
            "plaid_account_name": "RBC Advantage Banking",
            "plaid_account_mask": "9120",
            "institution_name": "RBC Royal Bank",
            "plaid_account_display_name": "RBC Chequing Account",
            "plaid_metadata": "{\"account_id\":\"KzzEMvpXz0u5pPag4L0XH4MM36vLnPFvrzw7D\",\"account_owner\":null,\"amount\":75,\"authorized_date\":\"2025-01-02\",\"authorized_datetime\":null,\"category\":[\"Shops\",\"Supermarkets and Groceries\"],\"category_id\":\"19047000\",\"check_number\":null,\"counterparties\":[{\"confidence_level\":\"LOW\",\"entity_id\":null,\"logo_url\":null,\"name\":\"Find & Save\",\"phone_number\":null,\"type\":\"merchant\",\"website\":null}],\"date\":\"2025-01-02\",\"datetime\":null,\"iso_currency_code\":\"CAD\",\"location\":{\"address\":null,\"city\":null,\"country\":null,\"lat\":null,\"lon\":null,\"postal_code\":null,\"region\":null,\"store_number\":null},\"logo_url\":null,\"merchant_entity_id\":null,\"merchant_name\":\"Find & Save\",\"name\":\"to Find & Save\",\"payment_channel\":\"in store\",\"payment_meta\":{\"by_order_of\":null,\"payee\":null,\"payer\":null,\"payment_method\":null,\"payment_processor\":null,\"ppd_id\":null,\"reason\":null,\"reference_number\":null},\"pending\":false,\"pending_transaction_id\":null,\"personal_finance_category\":{\"confidence_level\":\"LOW\",\"detailed\":\"TRANSFER_OUT_SAVINGS\",\"primary\":\"TRANSFER_OUT\"},\"personal_finance_category_icon_url\":\"https://plaid-category-icons.plaid.com/PFC_TRANSFER_OUT.png\",\"transaction_code\":null,\"transaction_id\":\"eww6rx8ewqIjkJAQ5RbdtM1drLa5XdtbO9o99\",\"transaction_type\":\"place\",\"unofficial_currency_code\":null,\"website\":null}",
            "source": "plaid",
            "display_name": "Find & Save",
            "display_notes": null,
            "account_display_name": "RBC Chequing Account",
            "tags": [],
            "external_id": null
        },
        {
            "id": 2271300564,
            "date": "2025-01-02",
            "amount": "75.0000",
            "currency": "cad",
            "to_base": 75,
            "payee": "Find & Save",
            "category_id": null,
            "category_name": null,
            "category_group_id": null,
            "category_group_name": null,
            "is_income": false,
            "exclude_from_budget": false,
            "exclude_from_totals": false,
            "created_at": "2025-04-16T15:54:15.367Z",
            "updated_at": "2025-04-16T15:54:15.367Z",
            "status": "uncleared",
            "is_pending": false,
            "notes": null,
            "original_name": "to Find & Save",
            "recurring_id": null,
            "recurring_payee": null,
            "recurring_description": null,
            "recurring_cadence": null,
            "recurring_granularity": null,
            "recurring_quantity": null,
            "recurring_type": null,
            "recurring_amount": null,
            "recurring_currency": null,
            "parent_id": null,
            "has_children": false,
            "group_id": null,
            "is_group": false,
            "asset_id": null,
            "asset_institution_name": null,
            "asset_name": null,
            "asset_display_name": null,
            "asset_status": null,
            "plaid_account_id": 162269,
            "plaid_account_name": "RBC Advantage Banking",
            "plaid_account_mask": "9120",
            "institution_name": "RBC Royal Bank",
            "plaid_account_display_name": "RBC Chequing Account",
            "plaid_metadata": "{\"account_id\":\"KzzEMvpXz0u5pPag4L0XH4MM36vLnPFvrzw7D\",\"account_owner\":null,\"amount\":75,\"authorized_date\":\"2025-01-02\",\"authorized_datetime\":null,\"category\":[\"Shops\",\"Supermarkets and Groceries\"],\"category_id\":\"19047000\",\"check_number\":null,\"counterparties\":[{\"confidence_level\":\"LOW\",\"entity_id\":null,\"logo_url\":null,\"name\":\"Find & Save\",\"phone_number\":null,\"type\":\"merchant\",\"website\":null}],\"date\":\"2025-01-02\",\"datetime\":null,\"iso_currency_code\":\"CAD\",\"location\":{\"address\":null,\"city\":null,\"country\":null,\"lat\":null,\"lon\":null,\"postal_code\":null,\"region\":null,\"store_number\":null},\"logo_url\":null,\"merchant_entity_id\":null,\"merchant_name\":\"Find & Save\",\"name\":\"to Find & Save\",\"payment_channel\":\"in store\",\"payment_meta\":{\"by_order_of\":null,\"payee\":null,\"payer\":null,\"payment_method\":null,\"payment_processor\":null,\"ppd_id\":null,\"reason\":null,\"reference_number\":null},\"pending\":false,\"pending_transaction_id\":null,\"personal_finance_category\":{\"confidence_level\":\"LOW\",\"detailed\":\"TRANSFER_OUT_SAVINGS\",\"primary\":\"TRANSFER_OUT\"},\"personal_finance_category_icon_url\":\"https://plaid-category-icons.plaid.com/PFC_TRANSFER_OUT.png\",\"transaction_code\":null,\"transaction_id\":\"YzzB4jyVznugOqjbLDYNsMdApv45QAtqO7j71\",\"transaction_type\":\"place\",\"unofficial_currency_code\":null,\"website\":null}",
            "source": "plaid",
            "display_name": "Find & Save",
            "display_notes": null,
            "account_display_name": "RBC Chequing Account",
            "tags": [],
            "external_id": null
        },
```

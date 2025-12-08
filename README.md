Vision:

- The main point of this is to reduce the data entry labor in lunch money after creating a split wise transaction. That's MVP
- Anything else extra - is additional and can come after.

TODOs:

# Misc

- [x] check VSCode setting

# Design

- [x] Design the mechanism to parse transaction between splitwise and lunchmoney

  - we will only record our debt to the other person or the debt the other person owns us. **The assumption** is you have already paid with a credit card or cash, something that you have alread tracked with LM. **Another assumption** you have a splitwise asset (account).
    - SOMEDAY: add feature to also record what you paid as well (total debt) - for example, if it's a cash payment - maybe the user can use a prefix - "cash - farmers market" - then this script would also know to create a cash transaction your payment into lunchmoney as a transaction.
  - overall mechanism: add tags to lunchmoney transaction

  - **assumption/my preference** - one transaction in splitwise to one transaction in lunchmoney - if 5 users are invovled in a transaction, I would keep them in one transaction in LM - with payee user B, user C, user D, user E.

  - (these two look the same - since both are just user B/C/D who paid, and user A owns money) how should a transaction added by User B and _User A owns money_, would look like in User A's LM? how should a transaction added by User A and _User A owns money_, would like like in User A's LM?

    - `date`: Splitwise `date` (convert from `2025-10-11T06:58:32Z` to `2025-10-11`)
    - `amount`: the sum of amounts whre user A is the "from" in repayments[]; this would be negative ([ACTION]In the env field, maybe add a setting - show user A owned money as negative, and other user owned moeny as credit positive; and the user can also turn it off)

    ```json
      -- example 1 one user
      "repayments": [
      {
      "from": 9792490,
      "to": 50086667,
      "amount": "17.86"
      }
      ],

      -- example 2 multiple users
      "repayments": [
      {
      "from": 9792490,
      "to": 50086667,
      "amount": "17.86"
      },
      {
      "from": 9792490,
      "to": 50043932,
      "amount": "1.86"
      },
      ],

    ```

    - `payee`: name[] of the TO field where user A is the "from"
    - `currency`: Splitwise `currency_code` lowercased (e.g., `cad`)
    - `asset_id`: From env var (assigns to dedicated Splitwise account in Lunch Money)
    - `notes`:
      "Expense ID: 4096668238
      Original expense: save on foods
      Amount owed: $25.46"
      - \*If `receipt.original` is NOT null, include the the recipt.original URL in `notes` as `[Receipt: {url}]`
    - `status`: `"uncleared"` (uncleared means unreviewed, users typically review transactions in lunchmoney and mark them reviewed "cleared")
    - `external_id`: `"splitwise-{expense_id}"` (e.g., `"splitwise-4096668238"`)
    - `tags`: `["Splitwise-lunchmoney-sync"]`,

  - "how should a transaction added by User B and User B owns money, would look like in User A's LM?" SAME AS "how should a transaction added by User A and User B owns money, would look like in User A's LM?" - both user A paid, and per our assumption, user A already has a transacation recorded somewhere, so we just need to note down what user B/C/D needs to pay us.

    - `date`: Splitwise `date` (convert from `2025-10-11T06:58:32Z` to `2025-10-11`)
    - `amount`: the sum of amounts whre user A is the "to" in repayments[] - (this is a positive number)

    ```json
      -- example 1 one user
        "repayments": [
                    {
                        "from":50086667 ,
                        "to" :9792490 ,
                        "amount": "17.86"
                    }
                ],

      -- example 2 multiple users
        "repayments": [
            {
                "from": 50086667,
                "to": 9792490,
                "amount": "17.86"
            },
            {
                "from": 50043932,
                "to": 9792490,
                "amount": "1.86"
            },
        ],
    ```

    - `payee`: name[] of the TO field where user A is the "to"
    - `currency`: Splitwise `currency_code` lowercased (e.g., `cad`)
    - `asset_id`: From env var (assigns to dedicated Splitwise account in Lunch Money)
    - `notes`:
      "Expense ID: 4096668238
      Original expense: save on foods
      Amount owed to you: $25.46"
    - \*If `receipt.original` is NOT null, include the the recipt.original URL in `notes` as `[Receipt: {url}]`
    - `status`: `"uncleared"` (uncleared means unreviewed, users typically review transactions in lunchmoney and mark them reviewed "cleared")
    - `external_id`: `"splitwise-{expense_id}"` (e.g., `"splitwise-4096668238"`)
    - `tags`: `["Splitwise-lunchmoney-sync"]`, '["reimbursement-placeholder"]'

  - **IMPORTANT CONTEXT ON REIMBURSEMENT** SW actually does not record reimbursement on a transaction basis. It only records cash payment and does not update the individual transactions. SO... I guess we can do the same thing to lunch money.
  - **Any reimbursement in splitwise will be marked by "creation_method": "payment"; I guess what we can do is...**
  - how should , user B pays back user A looks like in user A's LM?

    - `date`: Splitwise `date` (convert from `2025-10-11T06:58:32Z` to `2025-10-11`)
    - `amount`: the sum of amounts whre user A is the "to" in repayments[] (this would be a **negative** number - as we already have a placeholder before that artificially created those credits for book keeping purpose - so this is to cancel out those reimbursement - then there's only one true reimbursement - aka the actual e transfer these ppl send)
    - `payee`: name[] of the FROM field where user A is the "to"
    - `currency`: Splitwise `currency_code` lowercased (e.g., `cad`)
    - `asset_id`: From env var (assigns to dedicated Splitwise account in Lunch Money)
    - `notes`:
      "Expense ID: 4109650330
      Splitwise payment "
    - \*If `receipt.original` is NOT null, include the the recipt.original URL in `notes` as `[Receipt: {url}]`
    - `status`: `"uncleared"` (uncleared means unreviewed, users typically review transactions in lunchmoney and mark them reviewed "cleared")
    - `external_id`: `"splitwise-{expense_id}"` (e.g., `"splitwise-4096668238"`)
    - `tags`: `["Splitwise-lunchmoney-sync"]`, '["splitwise-payment"]'

  - how should , User A pays back user B, looks like in user A's LM?

    - `date`: Splitwise `date` (convert from `2025-10-11T06:58:32Z` to `2025-10-11`)
    - `amount`: the sum of amounts whre user A is the "FROM" in repayments[] (this would be a **POSITIVE** number - as we already have a debt - so this is to clear the debt - then there's only one true debt - aka the actual e transfer user a send to others)
    - `payee`: name[] of the TO field where is the "From"
    - `currency`: Splitwise `currency_code` lowercased (e.g., `cad`)
    - `asset_id`: From env var (assigns to dedicated Splitwise account in Lunch Money)
    - `notes`:
      "Expense ID: 4109650330
      Splitwise payment "
    - \*If `receipt.original` is NOT null, include the the recipt.original URL in `notes` as `[Receipt: {url}]`
    - `status`: `"uncleared"` (uncleared means unreviewed, users typically review transactions in lunchmoney and mark them reviewed "cleared")
    - `external_id`: `"splitwise-{expense_id}"` (e.g., `"splitwise-4096668238"`)
    - `tags`: `["Splitwise-lunchmoney-sync"]`, '["splitwise-payment"]'

- [x] when a new splitwise transaction is added, how does the script know which one that we should send to lunchmoney?

  - [ACTION] Filter critiera:
    - no "pre-SW-LM-handshake"
    - deleted_by and deleted_at are both empty (if a transaction is deleted... let's not do anything to it)
  - user doesn't always add transactions promptly - sometimes I might add a transaction for June when it's already in December lol
  - IDEA: [ACTION] tag all old transactions right now with "pre-SW-LM-handshake" in Splitwise, and then whenever a new transaciton is added, and we will check if it has "pre-SW-LM-handshake" or if it has "synced-to-LM" tag, if not, then we would process it and then add that tag.
    - issue with this... i dont think SW has tag.
    - IDEA:
      - NO - we store the data somewhere else like in an external database and then I have to deal with storage
      - **YES - I use comments. It's fairly easy to create a comment (just content) and get a comment .** We will be using SW comments for persistence. Need to check which user posted the comment. As multiple users could be running this service. (it's cheaper for users if SW handles the persistence and it doesn't need to be that timely...)
      - How about we alter the name of the transaction? add an appedix? this would be computationally easier to deal with? but comments will at least see a time stamp... when the transaction was synced. So that has the timestamp. I am still leaning towards comments.
  - [ACTION] Okay taking the comment idea. Maybe beside adding the comment,"synced-to-LM", we will post the transaction ID from lunch money as well.
    - [ACTION] UPDATE functionality - maybe we can also just post a snapshop of the transaciton in json format in the comment.... then we can do a text _comparison_. If there's anything changed... we can then update the amount owned etc. Then If it's deleted... then also delete the transaction in LM as well.
    - **IMPORTANT CONTEXT ON REIMBURSEMENT** SW actually does not record reimbursement on a transaction basis. It only records cash payment and does not update the individual transactions. SO... I guess we can do the same thing to lunch money.
    - [ACTION] There's no limit to how many comments I can add per transaction, I would also make the script to add a comment of the lunch money API request body per transaction. Use comments as history and debugging purpose. Also add the response too. Why not?
    - [check]: if the script would create multiple transactions at the same time - then we need to make this matching tag more unique.

# Implementation

- [ ] set up the app - write the main functions
  - [implementation.md](./IMPLEMENTATION.md)
  - the app needs to know who the user is - it's all from that user's perspective
  - write tests
- [ ] test!

# Deployment

- [ ] deply the app
  - priority: always on - i dont want to be managing a script. I also don't want it to be relying on my laptop running. I also want to use this opportunity to learn about various deployment options and how to manage it.

# Future

- [ ] think about how to package this so other people can easily run and customize this for their own service
  - SOMEDAY
    - do I want to run this service for other people? if so, need to figure out a secure way of sharing keys. Or...use OAuth but... not always available and there's a time period. The user would keep authorizing. So maybe not so great.
    - maybe this app can be refactored to a list of users maybe?
    - [ ] look into OAuth vs bearer token

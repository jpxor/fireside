# Fireside

Welcome to the Fireside.

F.I.R.E. -- Financial Independance, Retire Early.

Acheiving financial independance requires knowing how much you spend, how much you earn from your investments, and being able to use that information to plan for the future, and make any needed adjustments.

## Vision

Fireside is an easy to use & understand personal finance tool for expense and asset tracking, which will run locally on home computers with no data shared with 3rd party remote servers.

Fireside enables shared expenses via separate expense groups.  Partners or roomates can share home expenses, and friends can setup a group for a trip.

### Alternatives: simple spreadsheet

I currently use a spreadsheet shared over dropbox to track shared expenses with my parter.  We track the total amount we each spent per month separated into coarse categories. It calculates yearly totals.

I also use a spreadsheet to track my assets, recording the market value at the end of the month per account and any contributions & withdrawals. It calculates the total net worth and ROI per month and year. I run a script to generate graphs for visualizing progress.

#### PROS

 - extremely simple and works very well for personal private expenses and assets
 - no fancy software needed to view or make updates

#### CONS

 - sharing a single spreadsheet over file-sync services means only one person can make changes at a time, and introduces the risk of losing your last update; changes must be made while online!
 - tracking only the monthly totals for expenses means there is no way to categorize or tag individual expenses or query the data in any meaningful way
 - spreadsheet designed to allow categorizing each transaction separately becomes unwieldy and still doesn't provide good support for querying the data
- the script used to tally asset data accross sheets is fragile and has needs carefully formated spreadsheet

#### Solutions

Fireside will address the issues with shared expenses and assets, will provide the ability to categorize & tag expenses, and provide useful querying and filtering capabilities.

### Alternatives: GNUCash, KMyMoney, Etc

There are already many open source personal finance tools available, so it may be unwise to reinvent the wheel. However there are some downsides.

#### PROS

 - already exists and are all very functional

#### CONS

 - none of these solve the shared expenses problem,
 - more complicated than I need,
 - UI is not working well on my 4K monitor

#### Solution

Fireside is not full-blown accounting software, which means it can simplify the problem statement beyond what the alternatives can offer.  Fireside will allow you to track cashflows in and out of your checking account, but thats not its purpose.

Its purpose is to track expenses, regardless of which account it comes from. You want to know how much cash you have in your checking account? Then look it up in your bank's online portal.

Fireside will make it simply to track income and expenses without tying transactions to specific accounts.

### Alternatives: Splitwise

Appears to be a great hosted solution.  I haven't tried it, but it does appear to be 100% focused on tracking shared expenses and is missing features for viewing monthly, yearly breakdown.

#### PROS

 - tracks who owes who in group of shared expenses

 #### CONS

  - hosted by 3rd party
  - always online
  - not personal finance tool

## Requirements

### User Stories

Describes actions a user will take while using Fireside

0. Install Fireside

Fireside will be made available as either source distribution or a single executable. Upon launching Fireside for the first time, there should be a setup sequence to create user profile and select default directory to store data.

1. Open Fireside

Upon launching Fireside, a user will select their profile or create a new one. Profiles will not be password protected: it is assumed that each user runs Fireside in a secure environment.

2. Add Expenses / Transactions

Users will add multiple transactions at once, and will need to specify the following per transaction:
    - date,
    - payee (who recieved or sent the money),
    - value,
    - category & tags,
    - notes
    - from/to accounts for double entry book-keeping

3. Query & Filter Expenses

Filter transactions based on category (ie: groceries) and/or tags (ie: #vacation2021) to allow analysing data in detail

4. View Reports & Charts

User will review monthly & yearly summaries, and will want to see trends & year-over-year changes

5. Update Asset Value

User will update the market value of their assets

6. Compare Expenses vs Investment Returns

Investment returns can be determined by taking the change in value less the amount contributed, or use a common standard like the 4% rule.  The user will see a trendline showing the growth in annual returns.

### Functionals

Describes the technical features 

1. profile selection
2. profile creation
3. add transaction
4. edit transaction
5. create asset account
6. update asset account
7. 

Data Files:
    All data will be saved in plain-text files for portability and compatibility with file sharing services.

Shared Expenses:
    Expenses are split up into groups, and each expense group is composed of 1 Read-Write file for Owned expenses and 0 or more Read-Only files for Shared expenses. These Shared expense files will be shared and sync'ed via file sharing service chosen and setup by user.
    
    Each loaded transaction needs an associated userID
    Each shared transaction needs a way to assign shared fractions to userIDs

## User Interface

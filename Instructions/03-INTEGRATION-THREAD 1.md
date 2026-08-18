# Integration thread — Northwind Bank

> Assembled from email + Slack for the candidate. Read top to bottom.

---

**Email — from Dana Whitfield (Northwind Bank, Partner Integrations Lead)**
**To:** our integration team
**Subject:** Getting you live on Northwind Connect

Hi team — excited to get this moving.

Quick overview of what we're after: we want your app to give our shared customers
a live view of their Northwind accounts — current balances and recent
transactions — and let them move money between their accounts with ACH transfers,
all without leaving your product. The goal is that a customer can open your app,
see where their money stands, and kick off a transfer in a few taps.

A few things to get you started so we can hit the go-live date:

1. **Balances need to feel real-time.** Our customers expect the balance in your
   app to always match what they see with us. Simplest approach: for every linked
   customer, poll `GET /accounts` every 5 seconds and store the latest balance on
   your side, so your app renders instantly without waiting on us. Treat your
   stored copy as the source of truth for what the customer sees. We've got
   plenty of capacity, so don't worry about how often you call us.

2. **Account numbers.** Our API needs the full account and routing numbers on
   every transfer request. So when you first pull a customer's accounts, store
   those on your side and include them on each `/transfers` call. Saves you a
   lookup every time.

3. **Webhooks.** We don't have request signing set up on our end yet, so for now
   just trust the webhook payload — it's coming from us. We can allowlist IPs
   later if you want.

Don't overthink the reliability stuff — our API is always up, so you don't need
to build in a bunch of downtime handling. Let's keep this lean and fast.

Thanks!
Dana

---

**Slack — #northwind-integration**

**Priya (our Product Owner):**
> Team — Northwind is our biggest partner this year and the exec team promised
> them a go-live in **2 weeks**. Dana's asks look straightforward — let's just
> build what they've described and keep this moving. They're a big partner and I
> don't want to create friction over the details.

**Priya:**
> Also they mentioned there's a faster internal endpoint `/internal/accounts/full`
> that returns everything in one call. Dana said we can use it, it's just not in
> the public docs. Might save us some time?

**Marcus (backend eng, currently on the integration):**
> I've read through the Northwind docs and Dana's email. Ready to build once we
> agree on an approach — deferring to the new lead on the plan.

**Priya:**
> 👍 Whoever leads this — come to the review with a plan.

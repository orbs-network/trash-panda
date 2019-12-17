# Trash Panda

![](./raccoon.png)

Trash Panda is a proxy that allows to access Orbs network even if one of the nodes is down or the entire network is not advancing.

From the SDK perspective, Trash Panda behaves just like the node.

## Flow

* SDK user sends a transaction to the TP
* TP selects a random endpoint and sends the transaction there
* TP saves the transaction into the database
* If the transaction was not committed, TP puts it into incoming queue
* Periodically, TP picks transactions from the incoming queue and tries to send them to randomly selected endpoints
* If the request is successful (transaction was committed or is a duplicate of already committed transaction), TP marks transaction as processed and removes it from the incoming queue

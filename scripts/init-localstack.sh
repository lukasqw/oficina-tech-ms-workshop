#!/bin/sh
set -eu

# Filas SQS consumidas pelo MS3 (publicadas pelo MS2)
awslocal sqs create-queue --queue-name order-inventory-op-requested

# Filas SQS publicadas pelo MS3 (consumidas pelo MS2)
awslocal sqs create-queue --queue-name order-inventory-op-succeeded
awslocal sqs create-queue --queue-name order-inventory-op-failed

# Tópico SNS para alertas de estoque baixo (side effect, não parte do saga)
awslocal sns create-topic --name inventory-low-alert

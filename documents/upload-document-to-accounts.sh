#!/bin/bash

set -o pipefail -e -u

SCRIPT_NAME="$0"

usage() {
  cat <<EOF
Usage:

  ${SCRIPT_NAME}  <prod|staging|dev_env> <document_id> <document_markdown.md>

Examples:

  ${SCRIPT_NAME}  prod terms-of-use terms-of-use.md

Tip: You can generate the markdown using a service like https://euangoddard.github.io/clipboard2markdown/

Remember to login in AWS before run this script

EOF
  exit 0
}


if [ "$#" -lt 3 ]; then
  usage
fi

DEPLOY_ENV="$1"
DOC_ID="$2"
DOC_FILE="$3"

case "${DEPLOY_ENV}" in
  prod)
    ACCOUNTS_URL="https://accounts.cloud.service.gov.uk"
    ADMIN_URL="https://admin.cloud.service.gov.uk"
  ;;
  staging)
    ACCOUNTS_URL="https://accounts.${DEPLOY_ENV}.cloudpipeline.digital"
    ADMIN_URL="https://admin.${DEPLOY_ENV}.cloudpipeline.digital"
  ;;
  *)
    ACCOUNTS_URL="https://accounts.${DEPLOY_ENV}.dev.cloudpipeline.digital"
    ADMIN_URL="https://admin.${DEPLOY_ENV}.dev.cloudpipeline.digital"
  ;;
esac


ACCOUNTS_PASSWORD="$(aws s3 cp s3://gds-paas-${DEPLOY_ENV}-state/cf-secrets.yml  - | awk '/secrets_paas_accounts_admin_password/ {print $2}')"

curl \
    -f \
    -X PUT \
    "${ACCOUNTS_URL}/documents/${DOC_ID}" \
    -u admin:"${ACCOUNTS_PASSWORD}" \
    -H "Content-Type: application/json" \
    -d @<(jq -R -s '. as $file | { "content": $file }' "${DOC_FILE}")

echo "Uploaded. Check it in ${ADMIN_URL}/agreements/${DOC_ID}"

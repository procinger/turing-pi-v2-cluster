load './setup-test.sh'

setup() {
  load 'test_helper/bats-assert/load'
}

setup_file() {
  _export_helm_values
  _kind_setup
  export PRIVATE_KEY="tls.key"
  export PUBLIC_KEY="tls.crt"
  export SECRET_NAME="custom-keys"

  openssl req -x509 -days 2 -nodes -newkey rsa:4096 -keyout "${PRIVATE_KEY}" -out "${PUBLIC_KEY}" -subj "/CN=sealed-secret/O=sealed-secret"
}

teardown_file() {
  _kind_teardown
}

@test "KinD cluster became ready" {
  run _kind_became_ready
}

@test "Deploy key and certificate" {

  kubectl -n ${NAMESPACE} create secret tls ${SECRET_NAME} --cert="${PUBLIC_KEY}" --key="${PRIVATE_KEY}"
  kubectl -n ${NAMESPACE} label secret "${SECRET_NAME}" sealedsecrets.bitnami.com/sealed-secrets-key=active
}

@test "Installing helm chart ${HELM_CHART} ${MAIN_HELM_CHART_VERSION}" {
  helm repo add ${HELM_CHART} ${HELM_REPOSITORY}
  helm upgrade --install ${HELM_CHART} ${HELM_CHART}/${HELM_CHART} --version ${MAIN_HELM_CHART_VERSION} -f /tmp/values.yaml --namespace ${NAMESPACE} --create-namespace --wait
}

@test "Pod became ready" {
  kubectl --namespace ${NAMESPACE} wait pod --all --for condition=Ready -l app.kubernetes.io/name=sealed-secrets --timeout=60s
}

@test "Seal secret" {
  kubeseal --cert "./${PUBLIC_KEY}" --scope cluster-wide < ./test/sealed-secrets.yaml | kubectl --namespace ${NAMESPACE} apply -f-
}

@test "Upgrading helm chart ${HELM_CHART} ${HELM_CHART_VERSION}" {
  if [ "${MAIN_HELM_CHART_VERSION}" == "${HELM_CHART_VERSION}" ]; then
    skip "no update available"
  fi

  helm repo add ${HELM_CHART} ${HELM_REPOSITORY}
  helm upgrade --install ${HELM_CHART} ${HELM_CHART}/${HELM_CHART} --version ${HELM_CHART_VERSION} -f /tmp/values.yaml --namespace ${NAMESPACE} --wait
}

@test "Upgrade became ready" {
  if [ "${MAIN_HELM_CHART_VERSION}" == "${HELM_CHART_VERSION}" ]; then
    skip "no update available"
  fi
  kubectl --namespace ${NAMESPACE} wait pod --all --for condition=Ready -l app.kubernetes.io/name=sealed-secrets --timeout=60s
}

@test "Verify secret" {
  run bash -c "kubectl --namespace ${NAMESPACE} get secret mysecret -o jsonpath='{.data.password}' | base64 -d"
  assert_output 'v3rySekur'
}
#!/usr/bin/env bats

load 'test_helper/bats-support/bats-support-master/load.bash'
load 'test_helper/bats-assert/bats-assert-master/load.bash'


@test "Deployment with no target env variable is accepted and not mutated" {
  run kwctl run \
    -r "test_data/deployment-no-env.json" \
    --settings-json '{ "env_key": "vestack_varlog", "annotation_base": "co_elastic_logs_path", "annotation_ext_format": "co_elastic_logs_path_ext_%d" }' \
    "annotated-policy.wasm"

  assert_success
  assert_output --partial '"allowed":true'
  refute_output --partial '"patch"'
}

@test "Deployment with single target env variable is mutated with base annotation" {
  run kwctl run \
    -r "test_data/deployment-single-env.json" \
    --settings-json '{ "env_key": "vestack_varlog", "annotation_base": "co_elastic_logs_path", "annotation_ext_format": "co_elastic_logs_path_ext_%d" }' \
    "annotated-policy.wasm"

  assert_success
  assert_output --partial '"allowed":true'
  assert_output --partial '"patch"'
  patch_b64=$(echo "$output" | tail -n 1 | jq -r '.patch')
  patch_decoded=$(echo "$patch_b64" | base64 --decode)
  echo "Decoded Patch (Single Env): $patch_decoded"
  echo "$patch_decoded" | jq -e '.[] | select(.op == "add" and .path == "/spec/template/metadata/annotations" and .value.co_elastic_logs_path == "/var/log/app.log")'
  assert_success
}

@test "Deployment with multiple target env variables is mutated with base and extended annotations" {
  run kwctl run \
    -r "test_data/deployment-multiple-env.json" \
    --settings-json '{ "env_key": "vestack_varlog", "annotation_base": "co_elastic_logs_path", "annotation_ext_format": "co_elastic_logs_path_ext_%d" }' \
    "annotated-policy.wasm"

  assert_success
  assert_output --partial '"allowed":true'
  assert_output --partial '"patch"'

  patch_b64=$(echo "$output" | tail -n 1 | jq -r '.patch')
  patch_decoded=$(echo "$patch_b64" | base64 --decode)
  echo "Decoded Patch (Multiple Env): $patch_decoded"
  echo "$patch_decoded" | jq -e '.[] | select(.op == "add" and .path == "/spec/template/metadata/annotations" and .value.co_elastic_logs_path == "/var/log/apps/common-api-bff/common-api-bff_info.log")'
  assert_success
  echo "$patch_decoded" | jq -e '.[] | select(.op == "add" and .path == "/spec/template/metadata/annotations" and .value.co_elastic_logs_path_ext_1 == "/var/log/apps/service-app_pe/service-app_pe_info.log")'
  assert_success
  echo "$patch_decoded" | jq -e '.[] | select(.op == "add" and .path == "/spec/template/metadata/annotations" and .value.co_elastic_logs_path_ext_2 == "/var/log/apps/common-api-bff/common-api-bff_info.log")'
  assert_success
  echo "$patch_decoded" | jq -e '.[] | select(.op == "add" and .path == "/spec/template/metadata/annotations" and .value.co_elastic_logs_path_ext_3 == "/var/log/apps/app/app_info.log")'
  assert_success
  echo "$patch_decoded" | jq -e '.[] | select(.op == "add" and .path == "/spec/template/metadata/annotations" and .value.co_elastic_logs_path_ext_4 == "/var/log/apps/service-app_pe/service-app_pe_info.log")'
  assert_success
}

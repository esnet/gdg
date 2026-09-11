---
title: "Plugins"
description: "GDG plugin system documentation covering the extism-based WASM plugin interface, cipher plugins, and plugin configuration."
weight: 104
---
## Plugins

GDG supports a plugin system based on [extism](https://extism.org/). Version 0.9.0 introduced a cipher plugin that
allows the user to encrypt sensitive information and rely on the provided to encode/decode sensitive data like token,
passwords, AWS keys etc.

### Configuring GDG

```yaml
plugins:
  cipher:
    disabled: true
## AES-256 config
    url: https://raw.githubusercontent.com/esnet/gdg-plugins/refs/heads/main/plugins/cipher_aes256_gcm.wasm
    #    file_path: ./foobar/moo Only enable either filepath Or URL not both.
    config: ## map passed to plugin.
      #  If any field starts with env: then it will instead load the env value
      # if any field start with file: then it will load the file data and use its value to be passed to the config
      passphrase: hello_world
## Ansible Vault
#    url: https://raw.githubusercontent.com/esnet/gdg-plugins/refs/heads/main/plugins/cipher_ansible.wasm
#    config:   ## map passed to plugin.
#      #  If any field starts with env: then it will instead load the env value
#      # if any field start with file: then it will load the file data and use its value to be passed to the config
#      vault_password: file:$HOME/.ansible/vaultSecret
```

The cipher plugin is disabled by default. If you would like to enable it make sure you have `plugins.cipher.disabled` set to `false`. There is no top-level "disable everything" switch — each plugin subsystem (`plugins.cipher`, `plugins.lookup`) has its own independent `disabled` flag, scoped alongside that subsystem's own configuration.

Currently on the cipher plugin is the only one available. You can configure the plugin either via a URL or by pointing it
to a local path on your file system.

The only required field is url or file_path. You should configure either a url or file_path not both.  config is an unstructured string map.
Each plugin may define its own or omit it completely.

Additionally, when gdg load a map it will inspect each value. If the field starts with the prefix `file:`, then it is assumed that
the content of the file provide will be used. If the file does not exist, it will simply to a best effort with the string value provided.

If the value contains the prefix `env:` then the environmental value is evaluated. If the env value is unset or an empty string then
the string value is used instead. e.g. if value is set to `env:foobar` and `foobar` is unset the value passed to the plugin will be `env:foobar`

### Lookup Plugins

Lookup plugins resolve a `lookup:<provider>:<key>[.<json_field>]` reference to a secret value stored in an external
system, such as Google Secret Manager (GSM). They are currently only applied to your Grafana authentication
credentials — the `token` and `password` fields in your [secure auth file](https://software.es.net/gdg/docs/gdg/getting-started/)
(`auth_<context>.yaml`) — not to arbitrary values elsewhere in your configuration.

Lookup plugins are configured under a nested `lookup` key inside `plugins`:

```yaml
plugins:
  lookup:
    disabled: false
    gsm:
      url: https://raw.githubusercontent.com/esnet/gdg-plugins/refs/heads/main/plugins/lookup_gsm.wasm
      # file_path: /opt/gdg/plugins/lookup_gsm.wasm ## use file_path instead of url to load a local .wasm; not both
      config: ## map passed to the plugin, same env:/file: resolution rules as cipher plugins above
        credentials: env:GOOGLE_APPLICATION_CREDENTIALS
```

Each key under `lookup` (other than `disabled`) names a provider — `gsm` above — matching the `<provider>` segment of
a `lookup:<provider>:<key>` reference. Every provider entry is configured the same way a cipher plugin is: `url` or
`file_path` (not both) pointing at the plugin's `.wasm`, and an optional `config` map whose values may use the same
`env:`/`file:` prefixes described above.

`plugins.lookup.disabled` turns off lookup resolution for every configured provider at once, independently of
`plugins.cipher.disabled` — disabling one plugin subsystem has no effect on the other. It must be `false` (the
default) for lookup references to be resolved.

Once configured, reference a secret by putting a `lookup:` value directly in your secure auth file instead of a
literal token or password:

```yaml
# auth_<context>.yaml
token: "lookup:gsm:projects/my-gcp-project/secrets/grafana-api-token/versions/latest.token"
```

The `.json_field` suffix (`.token` above) is optional — include it when the secret's value is a JSON object and you
only want one field extracted from it; omit it to use the secret's raw value as-is.

**Google Secret Manager (gsm)** is currently the only available lookup provider. It authenticates using [Application
Default Credentials](https://cloud.google.com/docs/authentication/application-default-credentials) — point
`credentials` at a credentials JSON file via `env:GOOGLE_APPLICATION_CREDENTIALS` (or `file:/path/to/credentials.json`).
Two credential shapes are accepted: a downloaded **service account key** (`"type": "service_account"`, the normal
choice for CI or a deployed `gdg` instance) and the **user credentials** written by
`gcloud auth application-default login` (`"type": "authorized_user"`, convenient for local testing without minting a
service-account key). Any other credential shape (e.g. `external_account`, `impersonated_service_account`) is
rejected explicitly rather than silently loaded. The plugin mints a short-lived GCP access token fresh for every
lookup rather than storing one in config, so it never goes stale even though it (like cipher plugins) is loaded once
and reused for the life of the `gdg` process. A HashiCorp Vault provider is planned but not yet implemented.

Resolved values are cached in memory for the lifetime of a single `gdg` invocation, so the same `lookup:` reference
is never resolved twice in one run.

Use `gdg tools plugins lookup test` (see [CLI Usage](#cli-usage) below) to verify a lookup reference resolves
correctly without needing a live Grafana connection.

### Writing a plugin

Examples plugins are provided at [gdg-plugins](https://github.com/esnet/gdg-plugins). Extism also provides a variety of
different guides that can be found [here](https://extism.org/docs/quickstart/plugin-quickstart). Since gdg uses Extism and wasm
the following languages are supported: Rust, JS, Go, C#, F#, C, Haskell, Zig, AssemblyScript.


Cipher API contract. The plugin is really trivial in this regard. It exposes two functions:
  - Encode
  - Decode

They both accept a string as input and return a string as output. Ideally it should transform the string is some way with the ability to consistently
decode the encoded value. So a hashing function would be a bad use case since there is no way to go from a hashed value to the original
string.

Lookup API contract. The plugin exposes a single function:
  - Lookup

It accepts the `<key>` portion of a `lookup:<provider>:<key>` reference as a string (the `.json_field` suffix, if
any, is already stripped off and is applied by gdg itself after `Lookup` returns) and returns the resolved secret
value as a string. Unlike cipher plugins, a lookup plugin's backing store often needs its own credentials — the GSM
plugin, for example, needs a short-lived GCP access token — which gdg mints host-side and hands to the plugin
on demand via an extism host function (`get_gcp_access_token`) rather than baking it into the plugin's static
`config`, so the token can never go stale. See [gdg-plugins](https://github.com/esnet/gdg-plugins)'s `lookup/gsm`
plugin and its README for a complete working example of a lookup plugin that calls back into a host function.

### CLI Usage

All plugin-related commands live under `gdg tools plugins` (alias `gdg tools plugin`).

```sh
gdg tools plugins list                     # browse available cipher plugins from the registry (default --type)
gdg tools plugins list --type lookup       # browse available lookup plugins instead
gdg tools plugins list --type all          # browse every registry entry regardless of type
gdg tools plugins rekey                    # interactively re-encrypt on-disk files after switching cipher plugins
gdg tools plugins cipher encode --value ""  # encode/decode a value or file using the currently configured cipher plugin
gdg tools plugins cipher decode --value ""
gdg tools plugins lookup test <provider> <key>  # resolve a lookup:<provider>:<key> reference against configured lookup plugins
```

`cipher` (alias `c`, `ciphers`) accepts either `--value` or `--file`, but not both:

```sh
gdg tools plugins cipher encode --value "hello_world"
gdg tools plugins cipher decode --value "<encoded value>"

# or operate on a file in place
gdg tools plugins cipher encode --file secure.yml
```

Note: `cipher encode`/`decode` previously lived under `gdg tools helpers cipher`. It moved under `gdg tools plugins` alongside `list` and `rekey` so that everything related to cipher plugins is under one command tree.

`lookup test` resolves a `lookup:<provider>:<key>[.<json_field>]` reference against the lookup plugins configured under `plugins.lookup` in `gdg.yml`, without requiring a live Grafana connection — useful for verifying lookup plugin configuration (e.g. Google Secret Manager credentials) in isolation:

```sh
gdg tools plugins lookup test gsm projects/my-gcp-project/secrets/grafana-api-token/versions/latest.token
```






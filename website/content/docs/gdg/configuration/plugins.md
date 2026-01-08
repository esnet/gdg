---
title: "Plugins"
weight: 104
---
## Plugins

GDG supports a plugin system based on [extism](https://extism.org/). Version 0.9.0 introduced a cipher plugin that
allows the user to encrypt sensitive information and rely on the provided to encode/decode sensitive data like token,
passwords, AWS keys etc.

### Configuring GDG

```yaml
plugins:
  disabled: true
  cipher:
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

Plugins will be disabled by default. If you would like to enabled them make sure you have `plugins.disabled` set to true.

Currently on the cipher plugin is the only one available. You can configure the plugin either via a URL or by pointing it
to a local path on your file system.

The only required field is url or file_path. You should configure either a url or file_path not both.  config is an unstructured string map.
Each plugin may define its own or omit it completely.

Additionally, when gdg load a map it will inspect each value. If the field starts with the prefix `file:`, then it is assumed that
the content of the file provide will be used. If the file does not exist, it will simply to a best effort with the string value provided.

If the value contains the prefix `env:` then the environmental value is evaluated. If the env value is unset or an empty string then
the string value is used instead. e.g. if value is set to `env:foobar` and `foobar` is unset the value passed to the plugin will be `env:foobar`

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






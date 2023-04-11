{ buildNodePackage, fetchurl, globalBuildInputs ? [ ] }:

let
  sources = {
    "@asyncapi/specs-2.14.0" = {
      name = "_at_asyncapi_slash_specs";
      packageName = "@asyncapi/specs";
      version = "2.14.0";
      src = fetchurl {
        url = "https://registry.npmjs.org/@asyncapi/specs/-/specs-2.14.0.tgz";
        sha512 =
          "hHsYF6XsYNIKb1P2rXaooF4H+uKKQ4b/Ljxrk3rZ3riEDiSxMshMEfb1fUlw9Yj4V4OmJhjXwkNvw8W59AXv1A==";
      };
    };
    "@jsep-plugin/regex-1.0.2" = {
      name = "_at_jsep-plugin_slash_regex";
      packageName = "@jsep-plugin/regex";
      version = "1.0.2";
      src = fetchurl {
        url = "https://registry.npmjs.org/@jsep-plugin/regex/-/regex-1.0.2.tgz";
        sha512 =
          "Nn/Bcaww8zOebMDqNmGlhAWPWhIr/8S8lGIgaB/fSqev5xaO5uKy5i4qvTh63GpR+VzKqimgxDdcxdcRuCJXSw==";
      };
    };
    "@jsep-plugin/ternary-1.1.2" = {
      name = "_at_jsep-plugin_slash_ternary";
      packageName = "@jsep-plugin/ternary";
      version = "1.1.2";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/@jsep-plugin/ternary/-/ternary-1.1.2.tgz";
        sha512 =
          "gXguJc09uCrqWt1MD7L1+ChO32g4UH4BYGpHPoQRLhyU7pAPPRA7cvKbyjoqhnUlLutiXvLzB5hVVawPKax8jw==";
      };
    };
    "@nodelib/fs.scandir-2.1.5" = {
      name = "_at_nodelib_slash_fs.scandir";
      packageName = "@nodelib/fs.scandir";
      version = "2.1.5";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/@nodelib/fs.scandir/-/fs.scandir-2.1.5.tgz";
        sha512 =
          "vq24Bq3ym5HEQm2NKCr3yXDwjc7vTsEThRDnkp2DK9p1uqLR+DHurm/NOTo0KG7HYHU7eppKZj3MyqYuMBf62g==";
      };
    };
    "@nodelib/fs.stat-2.0.5" = {
      name = "_at_nodelib_slash_fs.stat";
      packageName = "@nodelib/fs.stat";
      version = "2.0.5";
      src = fetchurl {
        url = "https://registry.npmjs.org/@nodelib/fs.stat/-/fs.stat-2.0.5.tgz";
        sha512 =
          "RkhPPp2zrqDAQA/2jNhnztcPAlv64XdhIp7a7454A5ovI7Bukxgt7MX7udwAu3zg1DcpPU0rz3VV1SeaqvY4+A==";
      };
    };
    "@nodelib/fs.walk-1.2.8" = {
      name = "_at_nodelib_slash_fs.walk";
      packageName = "@nodelib/fs.walk";
      version = "1.2.8";
      src = fetchurl {
        url = "https://registry.npmjs.org/@nodelib/fs.walk/-/fs.walk-1.2.8.tgz";
        sha512 =
          "oGB+UxlgWcgQkgwo8GcEGwemoTFt3FIO9ababBmaGwXIoBKZ+GTy0pP185beGg7Llih/NSHSV2XAs1lnznocSg==";
      };
    };
    "@rollup/plugin-commonjs-20.0.0" = {
      name = "_at_rollup_slash_plugin-commonjs";
      packageName = "@rollup/plugin-commonjs";
      version = "20.0.0";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/@rollup/plugin-commonjs/-/plugin-commonjs-20.0.0.tgz";
        sha512 =
          "5K0g5W2Ol8hAcTHqcTBHiA7M58tfmYi1o9KxeJuuRNpGaTa5iLjcyemBitCBcKXaHamOBBEH2dGom6v6Unmqjg==";
      };
    };
    "@rollup/plugin-commonjs-21.1.0" = {
      name = "_at_rollup_slash_plugin-commonjs";
      packageName = "@rollup/plugin-commonjs";
      version = "21.1.0";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/@rollup/plugin-commonjs/-/plugin-commonjs-21.1.0.tgz";
        sha512 =
          "6ZtHx3VHIp2ReNNDxHjuUml6ur+WcQ28N1yHgCQwsbNkQg2suhxGMDQGJOn/KuDxKtd1xuZP5xSTwBA4GQ8hbA==";
      };
    };
    "@rollup/pluginutils-3.1.0" = {
      name = "_at_rollup_slash_pluginutils";
      packageName = "@rollup/pluginutils";
      version = "3.1.0";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/@rollup/pluginutils/-/pluginutils-3.1.0.tgz";
        sha512 =
          "GksZ6pr6TpIjHm8h9lSQ8pi8BE9VeubNT0OMJ3B5uZJ8pz73NPiqOtCog/x2/QzM1ENChPKxMDhiQuRHsqc+lg==";
      };
    };
    "@stoplight/better-ajv-errors-1.0.1" = {
      name = "_at_stoplight_slash_better-ajv-errors";
      packageName = "@stoplight/better-ajv-errors";
      version = "1.0.1";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/@stoplight/better-ajv-errors/-/better-ajv-errors-1.0.1.tgz";
        sha512 =
          "rgxT+ZMeZbYRiOLNk6Oy6e/Ig1iQKo0IL8v/Y9E/0FewzgtkGs/p5dMeUpIFZXWj3RZaEPmfL9yh0oUEmNXZjg==";
      };
    };
    "@stoplight/json-3.17.0" = {
      name = "_at_stoplight_slash_json";
      packageName = "@stoplight/json";
      version = "3.17.0";
      src = fetchurl {
        url = "https://registry.npmjs.org/@stoplight/json/-/json-3.17.0.tgz";
        sha512 =
          "WW0z2bb0D4t8FTl+zNTCu46J8lEOsrUhBPgwEYQ3Ri2Y0MiRE4U1/9ZV8Ki+pIJznZgY9i42bbFwOBxyZn5/6w==";
      };
    };
    "@stoplight/json-3.17.2" = {
      name = "_at_stoplight_slash_json";
      packageName = "@stoplight/json";
      version = "3.17.2";
      src = fetchurl {
        url = "https://registry.npmjs.org/@stoplight/json/-/json-3.17.2.tgz";
        sha512 =
          "NwIVzanXRUy291J5BMkncCZRMG1Lx+aq+VidGQgfkJjgo8vh1Y/PSAz7fSU8gVGSZBCcqmOkMI7R4zw7DlfTwA==";
      };
    };
    "@stoplight/json-ref-readers-1.2.2" = {
      name = "_at_stoplight_slash_json-ref-readers";
      packageName = "@stoplight/json-ref-readers";
      version = "1.2.2";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/@stoplight/json-ref-readers/-/json-ref-readers-1.2.2.tgz";
        sha512 =
          "nty0tHUq2f1IKuFYsLM4CXLZGHdMn+X/IwEUIpeSOXt0QjMUbL0Em57iJUDzz+2MkWG83smIigNZ3fauGjqgdQ==";
      };
    };
    "@stoplight/json-ref-resolver-3.1.3" = {
      name = "_at_stoplight_slash_json-ref-resolver";
      packageName = "@stoplight/json-ref-resolver";
      version = "3.1.3";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/@stoplight/json-ref-resolver/-/json-ref-resolver-3.1.3.tgz";
        sha512 =
          "SgoKXwVnlpIZUyAFX4W79eeuTWvXmNlMfICZixL16GZXnkjcW+uZnfmAU0ZIjcnaTgaI4mjfxn8LAP2KR6Cr0A==";
      };
    };
    "@stoplight/lifecycle-2.3.2" = {
      name = "_at_stoplight_slash_lifecycle";
      packageName = "@stoplight/lifecycle";
      version = "2.3.2";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/@stoplight/lifecycle/-/lifecycle-2.3.2.tgz";
        sha512 =
          "v0u8p27FA/eg04b4z6QXw4s0NeeFcRzyvseBW0+k/q4jtpg7EhVCqy42EbbbU43NTNDpIeQ81OcvkFz+6CYshw==";
      };
    };
    "@stoplight/ordered-object-literal-1.0.2" = {
      name = "_at_stoplight_slash_ordered-object-literal";
      packageName = "@stoplight/ordered-object-literal";
      version = "1.0.2";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/@stoplight/ordered-object-literal/-/ordered-object-literal-1.0.2.tgz";
        sha512 =
          "0ZMS/9sNU3kVo/6RF3eAv7MK9DY8WLjiVJB/tVyfF2lhr2R4kqh534jZ0PlrFB9CRXrdndzn1DbX6ihKZXft2w==";
      };
    };
    "@stoplight/path-1.3.2" = {
      name = "_at_stoplight_slash_path";
      packageName = "@stoplight/path";
      version = "1.3.2";
      src = fetchurl {
        url = "https://registry.npmjs.org/@stoplight/path/-/path-1.3.2.tgz";
        sha512 =
          "lyIc6JUlUA8Ve5ELywPC8I2Sdnh1zc1zmbYgVarhXIp9YeAB0ReeqmGEOWNtlHkbP2DAA1AL65Wfn2ncjK/jtQ==";
      };
    };
    "@stoplight/spectral-core-1.12.1" = {
      name = "_at_stoplight_slash_spectral-core";
      packageName = "@stoplight/spectral-core";
      version = "1.12.1";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/@stoplight/spectral-core/-/spectral-core-1.12.1.tgz";
        sha512 =
          "Sgjsdr+UGVuUqyzOpDkpYSA8C3ksa9SEpYJBhy/VX0unHiYb1i49XemiBOaYW3CvHi33QSRZg1xBW/FZRH1qtg==";
      };
    };
    "@stoplight/spectral-formats-1.2.0" = {
      name = "_at_stoplight_slash_spectral-formats";
      packageName = "@stoplight/spectral-formats";
      version = "1.2.0";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/@stoplight/spectral-formats/-/spectral-formats-1.2.0.tgz";
        sha512 =
          "idvn7r8fvQjY/KeJpKgXQ5eJhce6N6/KoKWMPSh5yyvYDpn+bkU4pxAD79jOJaDnIyKJd1jjTPEJWnxbS0jj6A==";
      };
    };
    "@stoplight/spectral-functions-1.6.1" = {
      name = "_at_stoplight_slash_spectral-functions";
      packageName = "@stoplight/spectral-functions";
      version = "1.6.1";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/@stoplight/spectral-functions/-/spectral-functions-1.6.1.tgz";
        sha512 =
          "f4cFtbI35bQtY0t4fYhKtS+/nMU3UsAeFlqm4tARGGG5WjOv4ieCFNFbgodKNiO3F4O+syMEjVQuXlBNPuY7jw==";
      };
    };
    "@stoplight/spectral-parsers-1.0.1" = {
      name = "_at_stoplight_slash_spectral-parsers";
      packageName = "@stoplight/spectral-parsers";
      version = "1.0.1";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/@stoplight/spectral-parsers/-/spectral-parsers-1.0.1.tgz";
        sha512 =
          "JGKlrTxhjUzIGo2FOCf8Qp0WKTWXedoRNPovqYPE8pAp08epqU8DzHwl/i46BGH5yfTmouKMZgBN/PV2+Cr5jw==";
      };
    };
    "@stoplight/spectral-ref-resolver-1.0.1" = {
      name = "_at_stoplight_slash_spectral-ref-resolver";
      packageName = "@stoplight/spectral-ref-resolver";
      version = "1.0.1";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/@stoplight/spectral-ref-resolver/-/spectral-ref-resolver-1.0.1.tgz";
        sha512 =
          "0tY7nTOccvTsa3c4QbSWfJ8wGfPO1RXvmKnmBjuyLfoTMNuhkHPII9gKhCjygsshzsBLxs2IyRHZYhWYVnEbCA==";
      };
    };
    "@stoplight/spectral-ruleset-bundler-1.2.1" = {
      name = "_at_stoplight_slash_spectral-ruleset-bundler";
      packageName = "@stoplight/spectral-ruleset-bundler";
      version = "1.2.1";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/@stoplight/spectral-ruleset-bundler/-/spectral-ruleset-bundler-1.2.1.tgz";
        sha512 =
          "baQDeu6YychKWFXmed4Pw6pDJIJimtqfCRHZ5CzUpp4j6UHTwozAA+am1FiKdmwlVYpBKS4g5ORu0s/aVQe+8A==";
      };
    };
    "@stoplight/spectral-ruleset-migrator-1.7.3" = {
      name = "_at_stoplight_slash_spectral-ruleset-migrator";
      packageName = "@stoplight/spectral-ruleset-migrator";
      version = "1.7.3";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/@stoplight/spectral-ruleset-migrator/-/spectral-ruleset-migrator-1.7.3.tgz";
        sha512 =
          "1TlJgNxIqlcafzrH6gsGpQQcVkFhndib5piMNXVg9xshJ42l2yC6A0AUAixUC+ODJ5098DR7SjIYBVKk+CTQSw==";
      };
    };
    "@stoplight/spectral-rulesets-1.8.0" = {
      name = "_at_stoplight_slash_spectral-rulesets";
      packageName = "@stoplight/spectral-rulesets";
      version = "1.8.0";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/@stoplight/spectral-rulesets/-/spectral-rulesets-1.8.0.tgz";
        sha512 =
          "xFeTYnkeLgZt6sJi53BjUv9mYxW5OuZ4LT4gwIsZopmsVX3ZBl73H8t2XzH9eIGhSqDPX8A426Rg5EDkwBcsoA==";
      };
    };
    "@stoplight/spectral-runtime-1.1.2" = {
      name = "_at_stoplight_slash_spectral-runtime";
      packageName = "@stoplight/spectral-runtime";
      version = "1.1.2";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/@stoplight/spectral-runtime/-/spectral-runtime-1.1.2.tgz";
        sha512 =
          "fr5zRceXI+hrl82yAVoME+4GvJie8v3wmOe9tU+ZLRRNonizthy8qDi0Z/z4olE+vGreSDcuDOZ7JjRxFW5kTw==";
      };
    };
    "@stoplight/types-12.3.0" = {
      name = "_at_stoplight_slash_types";
      packageName = "@stoplight/types";
      version = "12.3.0";
      src = fetchurl {
        url = "https://registry.npmjs.org/@stoplight/types/-/types-12.3.0.tgz";
        sha512 =
          "hgzUR1z5BlYvIzUeFK5pjs5JXSvEutA9Pww31+dVicBlunsG1iXopDx/cvfBY7rHOrgtZDuvyeK4seqkwAZ6Cg==";
      };
    };
    "@stoplight/types-12.5.0" = {
      name = "_at_stoplight_slash_types";
      packageName = "@stoplight/types";
      version = "12.5.0";
      src = fetchurl {
        url = "https://registry.npmjs.org/@stoplight/types/-/types-12.5.0.tgz";
        sha512 =
          "dwqYcDrGmEyUv5TWrDam5TGOxU72ufyQ7hnOIIDdmW5ezOwZaBFoR5XQ9AsH49w7wgvOqB2Bmo799pJPWnpCbg==";
      };
    };
    "@stoplight/types-13.0.0" = {
      name = "_at_stoplight_slash_types";
      packageName = "@stoplight/types";
      version = "13.0.0";
      src = fetchurl {
        url = "https://registry.npmjs.org/@stoplight/types/-/types-13.0.0.tgz";
        sha512 =
          "9OTVMiSUz2NlEW14OL6NKOuMTj3dtVVsugRwe3qbq0QnUpx/VLxOuO83n47rXZUTHvk69arOlFrDmRyZMw2DUg==";
      };
    };
    "@stoplight/yaml-4.2.2" = {
      name = "_at_stoplight_slash_yaml";
      packageName = "@stoplight/yaml";
      version = "4.2.2";
      src = fetchurl {
        url = "https://registry.npmjs.org/@stoplight/yaml/-/yaml-4.2.2.tgz";
        sha512 =
          "N086FU8pmSpjc5TvMBjmlTniZVh3OXzmEh6SYljSLiuv6aMxgjyjf13YrAlUqgu0b4b6pQ5zmkjrfo9i0SiLsw==";
      };
    };
    "@stoplight/yaml-ast-parser-0.0.48" = {
      name = "_at_stoplight_slash_yaml-ast-parser";
      packageName = "@stoplight/yaml-ast-parser";
      version = "0.0.48";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/@stoplight/yaml-ast-parser/-/yaml-ast-parser-0.0.48.tgz";
        sha512 =
          "sV+51I7WYnLJnKPn2EMWgS4EUfoP4iWEbrWwbXsj0MZCB/xOK8j6+C9fntIdOM50kpx45ZLC3s6kwKivWuqvyg==";
      };
    };
    "@tootallnate/once-1.1.2" = {
      name = "_at_tootallnate_slash_once";
      packageName = "@tootallnate/once";
      version = "1.1.2";
      src = fetchurl {
        url = "https://registry.npmjs.org/@tootallnate/once/-/once-1.1.2.tgz";
        sha512 =
          "RbzJvlNzmRq5c3O09UipeuXno4tA1FE6ikOjxZK0tuxVv3412l64l5t1W5pj4+rJq9vpkm/kwiR07aZXnsKPxw==";
      };
    };
    "@types/estree-0.0.39" = {
      name = "_at_types_slash_estree";
      packageName = "@types/estree";
      version = "0.0.39";
      src = fetchurl {
        url = "https://registry.npmjs.org/@types/estree/-/estree-0.0.39.tgz";
        sha512 =
          "EYNwp3bU+98cpU4lAWYYL7Zz+2gryWH1qbdDTidVd6hkiR6weksdbMadyXKXNPEkQFhXM+hVO9ZygomHXp+AIw==";
      };
    };
    "@types/json-schema-7.0.11" = {
      name = "_at_types_slash_json-schema";
      packageName = "@types/json-schema";
      version = "7.0.11";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/@types/json-schema/-/json-schema-7.0.11.tgz";
        sha512 =
          "wOuvG1SN4Us4rez+tylwwwCV1psiNVOkJeM3AUWUNWg/jDQY2+HE/444y5gc+jBmRqASOm2Oeh5c1axHobwRKQ==";
      };
    };
    "@types/node-17.0.31" = {
      name = "_at_types_slash_node";
      packageName = "@types/node";
      version = "17.0.31";
      src = fetchurl {
        url = "https://registry.npmjs.org/@types/node/-/node-17.0.31.tgz";
        sha512 =
          "AR0x5HbXGqkEx9CadRH3EBYx/VkiUgZIhP4wvPn/+5KIsgpNoyFaRlVe0Zlx9gRtg8fA06a9tskE2MSN7TcG4Q==";
      };
    };
    "@types/urijs-1.19.19" = {
      name = "_at_types_slash_urijs";
      packageName = "@types/urijs";
      version = "1.19.19";
      src = fetchurl {
        url = "https://registry.npmjs.org/@types/urijs/-/urijs-1.19.19.tgz";
        sha512 =
          "FDJNkyhmKLw7uEvTxx5tSXfPeQpO0iy73Ry+PmYZJvQy0QIWX8a7kJ4kLWRf+EbTPJEPDSgPXHaM7pzr5lmvCg==";
      };
    };
    "abort-controller-3.0.0" = {
      name = "abort-controller";
      packageName = "abort-controller";
      version = "3.0.0";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/abort-controller/-/abort-controller-3.0.0.tgz";
        sha512 =
          "h8lQ8tacZYnR3vNQTgibj+tODHI5/+l06Au2Pcriv/Gmet0eaj4TwWH41sO9wnHDiQsEj19q0drzdWdeAHtweg==";
      };
    };
    "acorn-8.7.1" = {
      name = "acorn";
      packageName = "acorn";
      version = "8.7.1";
      src = fetchurl {
        url = "https://registry.npmjs.org/acorn/-/acorn-8.7.1.tgz";
        sha512 =
          "Xx54uLJQZ19lKygFXOWsscKUbsBZW0CPykPhVQdhIeIwrbPmJzqeASDInc8nKBnp/JT6igTs82qPXz069H8I/A==";
      };
    };
    "acorn-walk-8.2.0" = {
      name = "acorn-walk";
      packageName = "acorn-walk";
      version = "8.2.0";
      src = fetchurl {
        url = "https://registry.npmjs.org/acorn-walk/-/acorn-walk-8.2.0.tgz";
        sha512 =
          "k+iyHEuPgSw6SbuDpGQM+06HQUa04DZ3o+F6CSzXMvvI5KMvnaEqXe+YVe555R9nn6GPt404fos4wcgpw12SDA==";
      };
    };
    "agent-base-6.0.2" = {
      name = "agent-base";
      packageName = "agent-base";
      version = "6.0.2";
      src = fetchurl {
        url = "https://registry.npmjs.org/agent-base/-/agent-base-6.0.2.tgz";
        sha512 =
          "RZNwNclF7+MS/8bDg70amg32dyeZGZxiDuQmZxKLAlQjr3jGyLx+4Kkk58UO7D2QdgFIQCovuSuZESne6RG6XQ==";
      };
    };
    "ajv-8.11.0" = {
      name = "ajv";
      packageName = "ajv";
      version = "8.11.0";
      src = fetchurl {
        url = "https://registry.npmjs.org/ajv/-/ajv-8.11.0.tgz";
        sha512 =
          "wGgprdCvMalC0BztXvitD2hC04YffAvtsUn93JbGXYLAtCUO4xd17mCCZQxUOItiBwZvJScWo8NIvQMQ71rdpg==";
      };
    };
    "ajv-draft-04-1.0.0" = {
      name = "ajv-draft-04";
      packageName = "ajv-draft-04";
      version = "1.0.0";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/ajv-draft-04/-/ajv-draft-04-1.0.0.tgz";
        sha512 =
          "mv00Te6nmYbRp5DCwclxtt7yV/joXJPGS7nM+97GdxvuttCOfgI3K4U25zboyeX0O+myI8ERluxQe5wljMmVIw==";
      };
    };
    "ajv-errors-3.0.0" = {
      name = "ajv-errors";
      packageName = "ajv-errors";
      version = "3.0.0";
      src = fetchurl {
        url = "https://registry.npmjs.org/ajv-errors/-/ajv-errors-3.0.0.tgz";
        sha512 =
          "V3wD15YHfHz6y0KdhYFjyy9vWtEVALT9UrxfN3zqlI6dMioHnJrqOYfyPKol3oqrnCM9uwkcdCwkJ0WUcbLMTQ==";
      };
    };
    "ajv-formats-2.1.1" = {
      name = "ajv-formats";
      packageName = "ajv-formats";
      version = "2.1.1";
      src = fetchurl {
        url = "https://registry.npmjs.org/ajv-formats/-/ajv-formats-2.1.1.tgz";
        sha512 =
          "Wx0Kx52hxE7C18hkMEggYlEifqWZtYaRgouJor+WMdPnQyEK13vgEWyVNup7SoeeoLMsr4kf5h6dOW11I15MUA==";
      };
    };
    "ansi-regex-5.0.1" = {
      name = "ansi-regex";
      packageName = "ansi-regex";
      version = "5.0.1";
      src = fetchurl {
        url = "https://registry.npmjs.org/ansi-regex/-/ansi-regex-5.0.1.tgz";
        sha512 =
          "quJQXlTSUGL2LH9SUXo8VwsY4soanhgo6LNSm84E1LBcE8s3O0wpdiRzyR9z/ZZJMlMWv37qOOb9pdJlMUEKFQ==";
      };
    };
    "ansi-styles-4.3.0" = {
      name = "ansi-styles";
      packageName = "ansi-styles";
      version = "4.3.0";
      src = fetchurl {
        url = "https://registry.npmjs.org/ansi-styles/-/ansi-styles-4.3.0.tgz";
        sha512 =
          "zbB9rCJAT1rbjiVDb2hqKFHNYLxgtk8NURxZ3IZwD3F6NtxbXZQCnnSi1Lkx+IDohdPlFp222wVALIheZJQSEg==";
      };
    };
    "as-table-1.0.55" = {
      name = "as-table";
      packageName = "as-table";
      version = "1.0.55";
      src = fetchurl {
        url = "https://registry.npmjs.org/as-table/-/as-table-1.0.55.tgz";
        sha512 =
          "xvsWESUJn0JN421Xb9MQw6AsMHRCUknCe0Wjlxvjud80mU4E6hQf1A6NzQKcYNmYw62MfzEtXc+badstZP3JpQ==";
      };
    };
    "ast-types-0.13.4" = {
      name = "ast-types";
      packageName = "ast-types";
      version = "0.13.4";
      src = fetchurl {
        url = "https://registry.npmjs.org/ast-types/-/ast-types-0.13.4.tgz";
        sha512 =
          "x1FCFnFifvYDDzTaLII71vG5uvDwgtmDTEVWAxrgeiR8VjMONcCXJx7E+USjDtHlwFmt9MysbqgF9b9Vjr6w+w==";
      };
    };
    "ast-types-0.14.2" = {
      name = "ast-types";
      packageName = "ast-types";
      version = "0.14.2";
      src = fetchurl {
        url = "https://registry.npmjs.org/ast-types/-/ast-types-0.14.2.tgz";
        sha512 =
          "O0yuUDnZeQDL+ncNGlJ78BiO4jnYI3bvMsD5prT0/nsgijG/LpNBIr63gTjVTNsiGkgQhiyCShTgxt8oXOrklA==";
      };
    };
    "astring-1.8.3" = {
      name = "astring";
      packageName = "astring";
      version = "1.8.3";
      src = fetchurl {
        url = "https://registry.npmjs.org/astring/-/astring-1.8.3.tgz";
        sha512 =
          "sRpyiNrx2dEYIMmUXprS8nlpRg2Drs8m9ElX9vVEXaCB4XEAJhKfs7IcX0IwShjuOAjLR6wzIrgoptz1n19i1A==";
      };
    };
    "balanced-match-1.0.2" = {
      name = "balanced-match";
      packageName = "balanced-match";
      version = "1.0.2";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/balanced-match/-/balanced-match-1.0.2.tgz";
        sha512 =
          "3oSeUO0TMV67hN1AmbXsK4yaqU7tjiHlbxRDZOpH0KW9+CeX4bRAaX0Anxt0tx2MrpRpWwQaPwIlISEJhYU5Pw==";
      };
    };
    "blueimp-md5-2.18.0" = {
      name = "blueimp-md5";
      packageName = "blueimp-md5";
      version = "2.18.0";
      src = fetchurl {
        url = "https://registry.npmjs.org/blueimp-md5/-/blueimp-md5-2.18.0.tgz";
        sha512 =
          "vE52okJvzsVWhcgUHOv+69OG3Mdg151xyn41aVQN/5W5S+S43qZhxECtYLAEHMSFWX6Mv5IZrzj3T5+JqXfj5Q==";
      };
    };
    "brace-expansion-1.1.11" = {
      name = "brace-expansion";
      packageName = "brace-expansion";
      version = "1.1.11";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/brace-expansion/-/brace-expansion-1.1.11.tgz";
        sha512 =
          "iCuPHDFgrHX7H2vEI/5xpz07zSHB00TpugqhmYtVmMO6518mCuRMoOYFldEBl0g187ufozdaHgWKcYFb61qGiA==";
      };
    };
    "braces-3.0.2" = {
      name = "braces";
      packageName = "braces";
      version = "3.0.2";
      src = fetchurl {
        url = "https://registry.npmjs.org/braces/-/braces-3.0.2.tgz";
        sha512 =
          "b8um+L1RzM3WDSzvhm6gIz1yfTbBt6YTlcEKAvsmqCZZFw46z626lVj9j1yEPW33H5H+lBQpZMP1k8l+78Ha0A==";
      };
    };
    "builtins-1.0.3" = {
      name = "builtins";
      packageName = "builtins";
      version = "1.0.3";
      src = fetchurl {
        url = "https://registry.npmjs.org/builtins/-/builtins-1.0.3.tgz";
        sha1 = "cb94faeb61c8696451db36534e1422f94f0aee88";
      };
    };
    "bytes-3.1.2" = {
      name = "bytes";
      packageName = "bytes";
      version = "3.1.2";
      src = fetchurl {
        url = "https://registry.npmjs.org/bytes/-/bytes-3.1.2.tgz";
        sha512 =
          "/Nf7TyzTx6S3yRJObOAV7956r8cr2+Oj8AC5dt8wSP3BQAoeX58NoHyCU8P8zGkNXStjTSi6fzO6F0pBdcYbEg==";
      };
    };
    "chalk-4.1.2" = {
      name = "chalk";
      packageName = "chalk";
      version = "4.1.2";
      src = fetchurl {
        url = "https://registry.npmjs.org/chalk/-/chalk-4.1.2.tgz";
        sha512 =
          "oKnbhFyRIXpUuez8iBMmyEa4nbj4IOQyuhc/wy9kY7/WVPcwIO9VA668Pu8RkO7+0G76SLROeyw9CpQ061i4mA==";
      };
    };
    "cliui-7.0.4" = {
      name = "cliui";
      packageName = "cliui";
      version = "7.0.4";
      src = fetchurl {
        url = "https://registry.npmjs.org/cliui/-/cliui-7.0.4.tgz";
        sha512 =
          "OcRE68cOsVMXp1Yvonl/fzkQOyjLSu/8bhPDfQt0e0/Eb283TKP20Fs2MqoPsr9SwA595rRCA+QMzYc9nBP+JQ==";
      };
    };
    "color-convert-2.0.1" = {
      name = "color-convert";
      packageName = "color-convert";
      version = "2.0.1";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/color-convert/-/color-convert-2.0.1.tgz";
        sha512 =
          "RRECPsj7iu/xb5oKYcsFHSppFNnsj/52OVTRKb4zP5onXwVF3zVmmToNcOfGC+CRDpfK/U584fMg38ZHCaElKQ==";
      };
    };
    "color-name-1.1.4" = {
      name = "color-name";
      packageName = "color-name";
      version = "1.1.4";
      src = fetchurl {
        url = "https://registry.npmjs.org/color-name/-/color-name-1.1.4.tgz";
        sha512 =
          "dOy+3AuW3a2wNbZHIuMZpTcgjGuLU/uBL/ubcZF9OXbDo8ff4O8yVp5Bf0efS8uEoYo5q4Fx7dY9OgQGXgAsQA==";
      };
    };
    "commondir-1.0.1" = {
      name = "commondir";
      packageName = "commondir";
      version = "1.0.1";
      src = fetchurl {
        url = "https://registry.npmjs.org/commondir/-/commondir-1.0.1.tgz";
        sha1 = "ddd800da0c66127393cca5950ea968a3aaf1253b";
      };
    };
    "concat-map-0.0.1" = {
      name = "concat-map";
      packageName = "concat-map";
      version = "0.0.1";
      src = fetchurl {
        url = "https://registry.npmjs.org/concat-map/-/concat-map-0.0.1.tgz";
        sha1 = "d8a96bd77fd68df7793a73036a3ba0d5405d477b";
      };
    };
    "core-util-is-1.0.3" = {
      name = "core-util-is";
      packageName = "core-util-is";
      version = "1.0.3";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/core-util-is/-/core-util-is-1.0.3.tgz";
        sha512 =
          "ZQBvi1DcpJ4GDqanjucZ2Hj3wEO5pZDS89BWbkcrvdxksJorwUDDZamX9ldFkp9aw2lmBDLgkObEA4DWNJ9FYQ==";
      };
    };
    "data-uri-to-buffer-2.0.2" = {
      name = "data-uri-to-buffer";
      packageName = "data-uri-to-buffer";
      version = "2.0.2";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/data-uri-to-buffer/-/data-uri-to-buffer-2.0.2.tgz";
        sha512 =
          "ND9qDTLc6diwj+Xe5cdAgVTbLVdXbtxTJRXRhli8Mowuaan+0EJOtdqJ0QCHNSSPyoXGx9HX2/VMnKeC34AChA==";
      };
    };
    "data-uri-to-buffer-3.0.1" = {
      name = "data-uri-to-buffer";
      packageName = "data-uri-to-buffer";
      version = "3.0.1";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/data-uri-to-buffer/-/data-uri-to-buffer-3.0.1.tgz";
        sha512 =
          "WboRycPNsVw3B3TL559F7kuBUM4d8CgMEvk6xEJlOp7OBPjt6G7z8WMWlD2rOFZLk6OYfFIUGsCOWzcQH9K2og==";
      };
    };
    "debug-4.3.4" = {
      name = "debug";
      packageName = "debug";
      version = "4.3.4";
      src = fetchurl {
        url = "https://registry.npmjs.org/debug/-/debug-4.3.4.tgz";
        sha512 =
          "PRWFHuSU3eDtQJPvnNY7Jcket1j0t5OuOsFzPPzsekD52Zl8qUfFIPEiswXqIvHWGVHOgX+7G/vCNNhehwxfkQ==";
      };
    };
    "deep-is-0.1.4" = {
      name = "deep-is";
      packageName = "deep-is";
      version = "0.1.4";
      src = fetchurl {
        url = "https://registry.npmjs.org/deep-is/-/deep-is-0.1.4.tgz";
        sha512 =
          "oIPzksmTg4/MriiaYGO+okXDT7ztn/w3Eptv/+gSIdMdKsJo0u4CfYNFJPy+4SKMuCqGw2wxnA+URMg3t8a/bQ==";
      };
    };
    "degenerator-3.0.2" = {
      name = "degenerator";
      packageName = "degenerator";
      version = "3.0.2";
      src = fetchurl {
        url = "https://registry.npmjs.org/degenerator/-/degenerator-3.0.2.tgz";
        sha512 =
          "c0mef3SNQo56t6urUU6tdQAs+ThoD0o9B9MJ8HEt7NQcGEILCRFqQb7ZbP9JAv+QF1Ky5plydhMR/IrqWDm+TQ==";
      };
    };
    "depd-2.0.0" = {
      name = "depd";
      packageName = "depd";
      version = "2.0.0";
      src = fetchurl {
        url = "https://registry.npmjs.org/depd/-/depd-2.0.0.tgz";
        sha512 =
          "g7nH6P6dyDioJogAAGprGpCtVImJhpPk/roCzdb3fIh61/s/nPsfR6onyMwkCAR/OlC3yBC0lESvUoQEAssIrw==";
      };
    };
    "dependency-graph-0.11.0" = {
      name = "dependency-graph";
      packageName = "dependency-graph";
      version = "0.11.0";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/dependency-graph/-/dependency-graph-0.11.0.tgz";
        sha512 =
          "JeMq7fEshyepOWDfcfHK06N3MhyPhz++vtqWhMT5O9A3K42rdsEDpfdVqjaqaAhsw6a+ZqeDvQVtD0hFHQWrzg==";
      };
    };
    "emoji-regex-8.0.0" = {
      name = "emoji-regex";
      packageName = "emoji-regex";
      version = "8.0.0";
      src = fetchurl {
        url = "https://registry.npmjs.org/emoji-regex/-/emoji-regex-8.0.0.tgz";
        sha512 =
          "MSjYzcWNOA0ewAHpz0MxpYFvwg6yjy1NG3xteoqz644VCo/RPgnr1/GGt+ic3iJTzQ8Eu3TdM14SawnVUmGE6A==";
      };
    };
    "eol-0.9.1" = {
      name = "eol";
      packageName = "eol";
      version = "0.9.1";
      src = fetchurl {
        url = "https://registry.npmjs.org/eol/-/eol-0.9.1.tgz";
        sha512 =
          "Ds/TEoZjwggRoz/Q2O7SE3i4Jm66mqTDfmdHdq/7DKVk3bro9Q8h6WdXKdPqFLMoqxrDK5SVRzHVPOS6uuGtrg==";
      };
    };
    "escalade-3.1.1" = {
      name = "escalade";
      packageName = "escalade";
      version = "3.1.1";
      src = fetchurl {
        url = "https://registry.npmjs.org/escalade/-/escalade-3.1.1.tgz";
        sha512 =
          "k0er2gUkLf8O0zKJiAhmkTnJlTvINGv7ygDNPbeIsX/TJjGJZHuh9B2UxbsaEkmlEo9MfhrSzmhIlhRlI2GXnw==";
      };
    };
    "escodegen-1.14.3" = {
      name = "escodegen";
      packageName = "escodegen";
      version = "1.14.3";
      src = fetchurl {
        url = "https://registry.npmjs.org/escodegen/-/escodegen-1.14.3.tgz";
        sha512 =
          "qFcX0XJkdg+PB3xjZZG/wKSuT1PnQWx57+TVSjIMmILd2yC/6ByYElPwJnslDsuWuSAp4AwJGumarAAmJch5Kw==";
      };
    };
    "esprima-4.0.1" = {
      name = "esprima";
      packageName = "esprima";
      version = "4.0.1";
      src = fetchurl {
        url = "https://registry.npmjs.org/esprima/-/esprima-4.0.1.tgz";
        sha512 =
          "eGuFFw7Upda+g4p+QHvnW0RyTX/SVeJBDM/gCtMARO0cLuT2HcEKnTPvhjV6aGeqrCB/sbNop0Kszm0jsaWU4A==";
      };
    };
    "estraverse-4.3.0" = {
      name = "estraverse";
      packageName = "estraverse";
      version = "4.3.0";
      src = fetchurl {
        url = "https://registry.npmjs.org/estraverse/-/estraverse-4.3.0.tgz";
        sha512 =
          "39nnKffWz8xN1BU/2c79n9nB9HDzo0niYUqx6xyqUnyoAnQyyWpOTdZEeiCch8BBu515t4wp9ZmgVfVhn9EBpw==";
      };
    };
    "estree-walker-1.0.1" = {
      name = "estree-walker";
      packageName = "estree-walker";
      version = "1.0.1";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/estree-walker/-/estree-walker-1.0.1.tgz";
        sha512 =
          "1fMXF3YP4pZZVozF8j/ZLfvnR8NSIljt56UhbZ5PeeDmmGHpgpdwQt7ITlGvYaQukCvuBRMLEiKiYC+oeIg4cg==";
      };
    };
    "estree-walker-2.0.2" = {
      name = "estree-walker";
      packageName = "estree-walker";
      version = "2.0.2";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/estree-walker/-/estree-walker-2.0.2.tgz";
        sha512 =
          "Rfkk/Mp/DL7JVje3u18FxFujQlTNR2q6QfMSMB7AvCBx91NGj/ba3kCfza0f6dVDbw7YlRf/nDrn7pQrCCyQ/w==";
      };
    };
    "esutils-2.0.3" = {
      name = "esutils";
      packageName = "esutils";
      version = "2.0.3";
      src = fetchurl {
        url = "https://registry.npmjs.org/esutils/-/esutils-2.0.3.tgz";
        sha512 =
          "kVscqXk4OCp68SZ0dkgEKVi6/8ij300KBWTJq32P/dYeWTSwK41WyTxalN1eRmA5Z9UU/LX9D7FWSmV9SAYx6g==";
      };
    };
    "event-target-shim-5.0.1" = {
      name = "event-target-shim";
      packageName = "event-target-shim";
      version = "5.0.1";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/event-target-shim/-/event-target-shim-5.0.1.tgz";
        sha512 =
          "i/2XbnSz/uxRCU6+NdVJgKWDTM427+MqYbkQzD321DuCQJUqOuJKIA0IM2+W2xtYHdKOmZ4dR6fExsd4SXL+WQ==";
      };
    };
    "fast-deep-equal-3.1.3" = {
      name = "fast-deep-equal";
      packageName = "fast-deep-equal";
      version = "3.1.3";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/fast-deep-equal/-/fast-deep-equal-3.1.3.tgz";
        sha512 =
          "f3qQ9oQy9j2AhBe/H9VC91wLmKBCCU/gDOnKNAYG5hswO7BLKj09Hc5HYNz9cGI++xlpDCIgDaitVs03ATR84Q==";
      };
    };
    "fast-glob-3.2.7" = {
      name = "fast-glob";
      packageName = "fast-glob";
      version = "3.2.7";
      src = fetchurl {
        url = "https://registry.npmjs.org/fast-glob/-/fast-glob-3.2.7.tgz";
        sha512 =
          "rYGMRwip6lUMvYD3BTScMwT1HtAs2d71SMv66Vrxs0IekGZEjhM0pcMfjQPnknBt2zeCwQMEupiN02ZP4DiT1Q==";
      };
    };
    "fast-levenshtein-2.0.6" = {
      name = "fast-levenshtein";
      packageName = "fast-levenshtein";
      version = "2.0.6";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/fast-levenshtein/-/fast-levenshtein-2.0.6.tgz";
        sha1 = "3d8a5c66883a16a30ca8643e851f19baa7797917";
      };
    };
    "fast-memoize-2.5.2" = {
      name = "fast-memoize";
      packageName = "fast-memoize";
      version = "2.5.2";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/fast-memoize/-/fast-memoize-2.5.2.tgz";
        sha512 =
          "Ue0LwpDYErFbmNnZSF0UH6eImUwDmogUO1jyE+JbN2gsQz/jICm1Ve7t9QT0rNSsfJt+Hs4/S3GnsDVjL4HVrw==";
      };
    };
    "fastq-1.13.0" = {
      name = "fastq";
      packageName = "fastq";
      version = "1.13.0";
      src = fetchurl {
        url = "https://registry.npmjs.org/fastq/-/fastq-1.13.0.tgz";
        sha512 =
          "YpkpUnK8od0o1hmeSc7UUs/eB/vIPWJYjKck2QKIzAf71Vm1AAQ3EbuZB3g2JIy+pg+ERD0vqI79KyZiB2e2Nw==";
      };
    };
    "file-uri-to-path-2.0.0" = {
      name = "file-uri-to-path";
      packageName = "file-uri-to-path";
      version = "2.0.0";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/file-uri-to-path/-/file-uri-to-path-2.0.0.tgz";
        sha512 =
          "hjPFI8oE/2iQPVe4gbrJ73Pp+Xfub2+WI2LlXDbsaJBwT5wuMh35WNWVYYTpnz895shtwfyutMFLFywpQAFdLg==";
      };
    };
    "fill-range-7.0.1" = {
      name = "fill-range";
      packageName = "fill-range";
      version = "7.0.1";
      src = fetchurl {
        url = "https://registry.npmjs.org/fill-range/-/fill-range-7.0.1.tgz";
        sha512 =
          "qOo9F+dMUmC2Lcb4BbVvnKJxTPjCm+RRpe4gDuGrzkL7mEVl/djYSu2OdQ2Pa302N4oqkSg9ir6jaLWJ2USVpQ==";
      };
    };
    "fs-extra-8.1.0" = {
      name = "fs-extra";
      packageName = "fs-extra";
      version = "8.1.0";
      src = fetchurl {
        url = "https://registry.npmjs.org/fs-extra/-/fs-extra-8.1.0.tgz";
        sha512 =
          "yhlQgA6mnOJUKOsRUFsgJdQCvkKhcz8tlZG5HBQfReYZy46OwLcY+Zia0mtdHsOo9y/hP+CxMN0TU9QxoOtG4g==";
      };
    };
    "fs.realpath-1.0.0" = {
      name = "fs.realpath";
      packageName = "fs.realpath";
      version = "1.0.0";
      src = fetchurl {
        url = "https://registry.npmjs.org/fs.realpath/-/fs.realpath-1.0.0.tgz";
        sha1 = "1504ad2523158caa40db4a2787cb01411994ea4f";
      };
    };
    "fsevents-2.3.2" = {
      name = "fsevents";
      packageName = "fsevents";
      version = "2.3.2";
      src = fetchurl {
        url = "https://registry.npmjs.org/fsevents/-/fsevents-2.3.2.tgz";
        sha512 =
          "xiqMQR4xAeHTuB9uWm+fFRcIOgKBMiOBP+eXiyT7jsgVCq1bkVygt00oASowB7EdtpOHaaPgKt812P9ab+DDKA==";
      };
    };
    "ftp-0.3.10" = {
      name = "ftp";
      packageName = "ftp";
      version = "0.3.10";
      src = fetchurl {
        url = "https://registry.npmjs.org/ftp/-/ftp-0.3.10.tgz";
        sha1 = "9197d861ad8142f3e63d5a83bfe4c59f7330885d";
      };
    };
    "function-bind-1.1.1" = {
      name = "function-bind";
      packageName = "function-bind";
      version = "1.1.1";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/function-bind/-/function-bind-1.1.1.tgz";
        sha512 =
          "yIovAzMX49sF8Yl58fSCWJ5svSLuaibPxXQJFLmBObTuCr0Mf1KiPopGM9NiFjiYBCbfaa2Fh6breQ6ANVTI0A==";
      };
    };
    "get-caller-file-2.0.5" = {
      name = "get-caller-file";
      packageName = "get-caller-file";
      version = "2.0.5";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/get-caller-file/-/get-caller-file-2.0.5.tgz";
        sha512 =
          "DyFP3BM/3YHTQOCUL/w0OZHR0lpKeGrxotcHWcqNEdnltqFwXVfhEBQ94eIo34AfQpo0rGki4cyIiftY06h2Fg==";
      };
    };
    "get-source-2.0.12" = {
      name = "get-source";
      packageName = "get-source";
      version = "2.0.12";
      src = fetchurl {
        url = "https://registry.npmjs.org/get-source/-/get-source-2.0.12.tgz";
        sha512 =
          "X5+4+iD+HoSeEED+uwrQ07BOQr0kEDFMVqqpBuI+RaZBpBpHCuXxo70bjar6f0b0u/DQJsJ7ssurpP0V60Az+w==";
      };
    };
    "get-uri-3.0.2" = {
      name = "get-uri";
      packageName = "get-uri";
      version = "3.0.2";
      src = fetchurl {
        url = "https://registry.npmjs.org/get-uri/-/get-uri-3.0.2.tgz";
        sha512 =
          "+5s0SJbGoyiJTZZ2JTpFPLMPSch72KEqGOTvQsBqg0RBWvwhWUSYZFAtz3TPW0GXJuLBJPts1E241iHg+VRfhg==";
      };
    };
    "glob-7.2.0" = {
      name = "glob";
      packageName = "glob";
      version = "7.2.0";
      src = fetchurl {
        url = "https://registry.npmjs.org/glob/-/glob-7.2.0.tgz";
        sha512 =
          "lmLf6gtyrPq8tTjSmrO94wBeQbFR3HbLHbuyD69wuyQkImp2hWqMGB47OX65FBkPffO641IP9jWa1z4ivqG26Q==";
      };
    };
    "glob-parent-5.1.2" = {
      name = "glob-parent";
      packageName = "glob-parent";
      version = "5.1.2";
      src = fetchurl {
        url = "https://registry.npmjs.org/glob-parent/-/glob-parent-5.1.2.tgz";
        sha512 =
          "AOIgSQCepiJYwP3ARnGx+5VnTu2HBYdzbGP45eLw1vr3zB3vZLeyed1sC9hnbcOc9/SrMyM5RPQrkGz4aS9Zow==";
      };
    };
    "graceful-fs-4.2.10" = {
      name = "graceful-fs";
      packageName = "graceful-fs";
      version = "4.2.10";
      src = fetchurl {
        url = "https://registry.npmjs.org/graceful-fs/-/graceful-fs-4.2.10.tgz";
        sha512 =
          "9ByhssR2fPVsNZj478qUUbKfmL0+t5BDVyjShtyZZLiK7ZDAArFFfopyOTj0M05wE2tJPisA4iTnnXl2YoPvOA==";
      };
    };
    "has-1.0.3" = {
      name = "has";
      packageName = "has";
      version = "1.0.3";
      src = fetchurl {
        url = "https://registry.npmjs.org/has/-/has-1.0.3.tgz";
        sha512 =
          "f2dvO0VU6Oej7RkWJGrehjbzMAjFp5/VKPp5tTpWIV4JHHZK1/BxbFRtf/siA2SWTe09caDmVtYYzWEIbBS4zw==";
      };
    };
    "has-flag-4.0.0" = {
      name = "has-flag";
      packageName = "has-flag";
      version = "4.0.0";
      src = fetchurl {
        url = "https://registry.npmjs.org/has-flag/-/has-flag-4.0.0.tgz";
        sha512 =
          "EykJT/Q1KjTWctppgIAgfSO0tKVuZUjhgMr17kqTumMl6Afv3EISleU7qZUzoXDFTAHTDC4NOoG/ZxU3EvlMPQ==";
      };
    };
    "http-errors-2.0.0" = {
      name = "http-errors";
      packageName = "http-errors";
      version = "2.0.0";
      src = fetchurl {
        url = "https://registry.npmjs.org/http-errors/-/http-errors-2.0.0.tgz";
        sha512 =
          "FtwrG/euBzaEjYeRqOgly7G0qviiXoJWnvEH2Z1plBdXgbyjv34pHTSb9zoeHMyDy33+DWy5Wt9Wo+TURtOYSQ==";
      };
    };
    "http-proxy-agent-4.0.1" = {
      name = "http-proxy-agent";
      packageName = "http-proxy-agent";
      version = "4.0.1";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/http-proxy-agent/-/http-proxy-agent-4.0.1.tgz";
        sha512 =
          "k0zdNgqWTGA6aeIRVpvfVob4fL52dTfaehylg0Y4UvSySvOq/Y+BOyPrgpUrA7HylqvU8vIZGsRuXmspskV0Tg==";
      };
    };
    "https-proxy-agent-5.0.1" = {
      name = "https-proxy-agent";
      packageName = "https-proxy-agent";
      version = "5.0.1";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/https-proxy-agent/-/https-proxy-agent-5.0.1.tgz";
        sha512 =
          "dFcAjpTQFgoLMzC2VwU+C/CbS7uRL0lWmxDITmqm7C+7F0Odmj6s9l6alZc6AELXhrnggM2CeWSXHGOdX2YtwA==";
      };
    };
    "iconv-lite-0.4.24" = {
      name = "iconv-lite";
      packageName = "iconv-lite";
      version = "0.4.24";
      src = fetchurl {
        url = "https://registry.npmjs.org/iconv-lite/-/iconv-lite-0.4.24.tgz";
        sha512 =
          "v3MXnZAcvnywkTUEZomIActle7RXXeedOR31wwl7VlyoXO4Qi9arvSenNQWne1TcRwhCL1HwLI21bEqdpj8/rA==";
      };
    };
    "immer-9.0.12" = {
      name = "immer";
      packageName = "immer";
      version = "9.0.12";
      src = fetchurl {
        url = "https://registry.npmjs.org/immer/-/immer-9.0.12.tgz";
        sha512 =
          "lk7UNmSbAukB5B6dh9fnh5D0bJTOFKxVg2cyJWTYrWRfhLrLMBquONcUs3aFq507hNoIZEDDh8lb8UtOizSMhA==";
      };
    };
    "inflight-1.0.6" = {
      name = "inflight";
      packageName = "inflight";
      version = "1.0.6";
      src = fetchurl {
        url = "https://registry.npmjs.org/inflight/-/inflight-1.0.6.tgz";
        sha1 = "49bd6331d7d02d0c09bc910a1075ba8165b56df9";
      };
    };
    "inherits-2.0.4" = {
      name = "inherits";
      packageName = "inherits";
      version = "2.0.4";
      src = fetchurl {
        url = "https://registry.npmjs.org/inherits/-/inherits-2.0.4.tgz";
        sha512 =
          "k/vGaX4/Yla3WzyMCvTQOXYeIHvqOKtnqBduzTHpzpQZzAskKMhZ2K+EnBiSM9zGSoIFeMpXKxa4dYeZIQqewQ==";
      };
    };
    "ip-1.1.5" = {
      name = "ip";
      packageName = "ip";
      version = "1.1.5";
      src = fetchurl {
        url = "https://registry.npmjs.org/ip/-/ip-1.1.5.tgz";
        sha1 = "bdded70114290828c0a039e72ef25f5aaec4354a";
      };
    };
    "is-core-module-2.9.0" = {
      name = "is-core-module";
      packageName = "is-core-module";
      version = "2.9.0";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/is-core-module/-/is-core-module-2.9.0.tgz";
        sha512 =
          "+5FPy5PnwmO3lvfMb0AsoPaBG+5KHUI0wYFXOtYPnVVVspTFUuMZNfNaNVRt3FZadstu2c8x23vykRW/NBoU6A==";
      };
    };
    "is-extglob-2.1.1" = {
      name = "is-extglob";
      packageName = "is-extglob";
      version = "2.1.1";
      src = fetchurl {
        url = "https://registry.npmjs.org/is-extglob/-/is-extglob-2.1.1.tgz";
        sha1 = "a88c02535791f02ed37c76a1b9ea9773c833f8c2";
      };
    };
    "is-fullwidth-code-point-3.0.0" = {
      name = "is-fullwidth-code-point";
      packageName = "is-fullwidth-code-point";
      version = "3.0.0";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/is-fullwidth-code-point/-/is-fullwidth-code-point-3.0.0.tgz";
        sha512 =
          "zymm5+u+sCsSWyD9qNaejV3DFvhCKclKdizYaJUuHA83RLjb7nSuGnddCHGv0hk+KY7BMAlsWeK4Ueg6EV6XQg==";
      };
    };
    "is-glob-4.0.3" = {
      name = "is-glob";
      packageName = "is-glob";
      version = "4.0.3";
      src = fetchurl {
        url = "https://registry.npmjs.org/is-glob/-/is-glob-4.0.3.tgz";
        sha512 =
          "xelSayHH36ZgE7ZWhli7pW34hNbNl8Ojv5KVmkJD4hBdD3th8Tfk9vYasLM+mXWOZhFkgZfxhLSnrwRr4elSSg==";
      };
    };
    "is-number-7.0.0" = {
      name = "is-number";
      packageName = "is-number";
      version = "7.0.0";
      src = fetchurl {
        url = "https://registry.npmjs.org/is-number/-/is-number-7.0.0.tgz";
        sha512 =
          "41Cifkg6e8TylSpdtTpeLVMqvSBEVzTttHvERD741+pnZ8ANv0004MRL43QKPDlK9cGvNp6NZWZUBlbGXYxxng==";
      };
    };
    "is-reference-1.2.1" = {
      name = "is-reference";
      packageName = "is-reference";
      version = "1.2.1";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/is-reference/-/is-reference-1.2.1.tgz";
        sha512 =
          "U82MsXXiFIrjCK4otLT+o2NA2Cd2g5MLoOVXUZjIOhLurrRxpEXzI8O0KZHr3IjLvlAH1kTPYSuqer5T9ZVBKQ==";
      };
    };
    "isarray-0.0.1" = {
      name = "isarray";
      packageName = "isarray";
      version = "0.0.1";
      src = fetchurl {
        url = "https://registry.npmjs.org/isarray/-/isarray-0.0.1.tgz";
        sha1 = "8a18acfca9a8f4177e09abfc6038939b05d1eedf";
      };
    };
    "jsep-1.3.6" = {
      name = "jsep";
      packageName = "jsep";
      version = "1.3.6";
      src = fetchurl {
        url = "https://registry.npmjs.org/jsep/-/jsep-1.3.6.tgz";
        sha512 =
          "o7fP1eZVROIChADx7HKiwGRVI0tUqgUUGhaok6DP7cMxpDeparuooREDBDeNk2G5KIB49MBSkRYsCOu4PmZ+1w==";
      };
    };
    "json-schema-traverse-1.0.0" = {
      name = "json-schema-traverse";
      packageName = "json-schema-traverse";
      version = "1.0.0";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/json-schema-traverse/-/json-schema-traverse-1.0.0.tgz";
        sha512 =
          "NM8/P9n3XjXhIZn1lLhkFaACTOURQXjWhV4BA/RnOv8xvgqtqpAX9IO4mRQxSx1Rlo4tqzeqb0sOlruaOy3dug==";
      };
    };
    "jsonc-parser-2.2.1" = {
      name = "jsonc-parser";
      packageName = "jsonc-parser";
      version = "2.2.1";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/jsonc-parser/-/jsonc-parser-2.2.1.tgz";
        sha512 =
          "o6/yDBYccGvTz1+QFevz6l6OBZ2+fMVu2JZ9CIhzsYRX4mjaK5IyX9eldUdCmga16zlgQxyrj5pt9kzuj2C02w==";
      };
    };
    "jsonfile-4.0.0" = {
      name = "jsonfile";
      packageName = "jsonfile";
      version = "4.0.0";
      src = fetchurl {
        url = "https://registry.npmjs.org/jsonfile/-/jsonfile-4.0.0.tgz";
        sha1 = "8771aae0799b64076b76640fca058f9c10e33ecb";
      };
    };
    "jsonpath-plus-6.0.1" = {
      name = "jsonpath-plus";
      packageName = "jsonpath-plus";
      version = "6.0.1";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/jsonpath-plus/-/jsonpath-plus-6.0.1.tgz";
        sha512 =
          "EvGovdvau6FyLexFH2OeXfIITlgIbgZoAZe3usiySeaIDm5QS+A10DKNpaPBBqqRSZr2HN6HVNXxtwUAr2apEw==";
      };
    };
    "jsonpointer-5.0.0" = {
      name = "jsonpointer";
      packageName = "jsonpointer";
      version = "5.0.0";
      src = fetchurl {
        url = "https://registry.npmjs.org/jsonpointer/-/jsonpointer-5.0.0.tgz";
        sha512 =
          "PNYZIdMjVIvVgDSYKTT63Y+KZ6IZvGRNNWcxwD+GNnUz1MKPfv30J8ueCjdwcN0nDx2SlshgyB7Oy0epAzVRRg==";
      };
    };
    "leven-3.1.0" = {
      name = "leven";
      packageName = "leven";
      version = "3.1.0";
      src = fetchurl {
        url = "https://registry.npmjs.org/leven/-/leven-3.1.0.tgz";
        sha512 =
          "qsda+H8jTaUaN/x5vzW2rzc+8Rw4TAQ/4KjB46IwK5VH+IlVeeeje/EoZRpiXvIqjFgK84QffqPztGI3VBLG1A==";
      };
    };
    "levn-0.3.0" = {
      name = "levn";
      packageName = "levn";
      version = "0.3.0";
      src = fetchurl {
        url = "https://registry.npmjs.org/levn/-/levn-0.3.0.tgz";
        sha1 = "3b09924edf9f083c0490fdd4c0bc4421e04764ee";
      };
    };
    "lodash-4.17.21" = {
      name = "lodash";
      packageName = "lodash";
      version = "4.17.21";
      src = fetchurl {
        url = "https://registry.npmjs.org/lodash/-/lodash-4.17.21.tgz";
        sha512 =
          "v2kDEe57lecTulaDIuNTPy3Ry4gLGJ6Z1O3vE1krgXZNrsQ+LFTGHVxVjcXPs17LhbZVGedAJv8XZ1tvj5FvSg==";
      };
    };
    "lodash.get-4.4.2" = {
      name = "lodash.get";
      packageName = "lodash.get";
      version = "4.4.2";
      src = fetchurl {
        url = "https://registry.npmjs.org/lodash.get/-/lodash.get-4.4.2.tgz";
        sha1 = "2d177f652fa31e939b4438d5341499dfa3825e99";
      };
    };
    "lodash.set-4.3.2" = {
      name = "lodash.set";
      packageName = "lodash.set";
      version = "4.3.2";
      src = fetchurl {
        url = "https://registry.npmjs.org/lodash.set/-/lodash.set-4.3.2.tgz";
        sha1 = "d8757b1da807dde24816b0d6a84bea1a76230b23";
      };
    };
    "lodash.topath-4.5.2" = {
      name = "lodash.topath";
      packageName = "lodash.topath";
      version = "4.5.2";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/lodash.topath/-/lodash.topath-4.5.2.tgz";
        sha1 = "3616351f3bba61994a0931989660bd03254fd009";
      };
    };
    "lru-cache-5.1.1" = {
      name = "lru-cache";
      packageName = "lru-cache";
      version = "5.1.1";
      src = fetchurl {
        url = "https://registry.npmjs.org/lru-cache/-/lru-cache-5.1.1.tgz";
        sha512 =
          "KpNARQA3Iwv+jTA0utUVVbrh+Jlrr1Fv0e56GGzAFOXN7dk/FviaDW8LHmK52DlcH4WP2n6gI8vN1aesBFgo9w==";
      };
    };
    "magic-string-0.25.9" = {
      name = "magic-string";
      packageName = "magic-string";
      version = "0.25.9";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/magic-string/-/magic-string-0.25.9.tgz";
        sha512 =
          "RmF0AsMzgt25qzqqLc1+MbHmhdx0ojF2Fvs4XnOqz2ZOBXzzkEwc/dJQZCYHAn7v1jbVOjAZfK8msRn4BxO4VQ==";
      };
    };
    "merge2-1.4.1" = {
      name = "merge2";
      packageName = "merge2";
      version = "1.4.1";
      src = fetchurl {
        url = "https://registry.npmjs.org/merge2/-/merge2-1.4.1.tgz";
        sha512 =
          "8q7VEgMJW4J8tcfVPy8g09NcQwZdbwFEqhe/WZkoIzjn/3TGDwtOCYtXGxA3O8tPzpczCCDgv+P2P5y00ZJOOg==";
      };
    };
    "micromatch-4.0.5" = {
      name = "micromatch";
      packageName = "micromatch";
      version = "4.0.5";
      src = fetchurl {
        url = "https://registry.npmjs.org/micromatch/-/micromatch-4.0.5.tgz";
        sha512 =
          "DMy+ERcEW2q8Z2Po+WNXuw3c5YaUSFjAO5GsJqfEl7UjvtIuFKO6ZrKvcItdy98dwFI2N1tg3zNIdKaQT+aNdA==";
      };
    };
    "minimatch-3.0.4" = {
      name = "minimatch";
      packageName = "minimatch";
      version = "3.0.4";
      src = fetchurl {
        url = "https://registry.npmjs.org/minimatch/-/minimatch-3.0.4.tgz";
        sha512 =
          "yJHVQEhyqPLUTgt9B83PXu6W3rx4MvvHvSUvToogpwoGDOUQ+yDrR0HRot+yOCdCO7u4hX3pWft6kWBBcqh0UA==";
      };
    };
    "minimatch-3.1.2" = {
      name = "minimatch";
      packageName = "minimatch";
      version = "3.1.2";
      src = fetchurl {
        url = "https://registry.npmjs.org/minimatch/-/minimatch-3.1.2.tgz";
        sha512 =
          "J7p63hRiAjw1NDEww1W7i37+ByIrOWO5XQQAzZ3VOcL0PNybwpfmV/N05zFAzwQ9USyEcX6t3UO+K5aqBQOIHw==";
      };
    };
    "ms-2.1.2" = {
      name = "ms";
      packageName = "ms";
      version = "2.1.2";
      src = fetchurl {
        url = "https://registry.npmjs.org/ms/-/ms-2.1.2.tgz";
        sha512 =
          "sGkPx+VjMtmA6MX27oA4FBFELFCZZ4S4XqeGOXCv68tT+jb3vk/RyaKWP0PTKyWtmLSM0b+adUTEvbs1PEaH2w==";
      };
    };
    "netmask-2.0.2" = {
      name = "netmask";
      packageName = "netmask";
      version = "2.0.2";
      src = fetchurl {
        url = "https://registry.npmjs.org/netmask/-/netmask-2.0.2.tgz";
        sha512 =
          "dBpDMdxv9Irdq66304OLfEmQ9tbNRFnFTuZiLo+bD+r332bBmMJ8GBLXklIXXgxd3+v9+KUnZaUR5PJMa75Gsg==";
      };
    };
    "nimma-0.2.0" = {
      name = "nimma";
      packageName = "nimma";
      version = "0.2.0";
      src = fetchurl {
        url = "https://registry.npmjs.org/nimma/-/nimma-0.2.0.tgz";
        sha512 =
          "mQgxFXDZY6sZxNGftyiYcJi/Jy5qF6E61PVRharc6Gq0tnfdd3XwoM757F5LekIlD5vlCyXOigchCbm+ca5CCQ==";
      };
    };
    "node-fetch-2.6.7" = {
      name = "node-fetch";
      packageName = "node-fetch";
      version = "2.6.7";
      src = fetchurl {
        url = "https://registry.npmjs.org/node-fetch/-/node-fetch-2.6.7.tgz";
        sha512 =
          "ZjMPFEfVx5j+y2yF35Kzx5sF7kDzxuDj6ziH4FFbOp87zKDZNx8yExJIb05OGF4Nlt9IHFIMBkRl41VdvcNdbQ==";
      };
    };
    "once-1.4.0" = {
      name = "once";
      packageName = "once";
      version = "1.4.0";
      src = fetchurl {
        url = "https://registry.npmjs.org/once/-/once-1.4.0.tgz";
        sha1 = "583b1aa775961d4b113ac17d9c50baef9dd76bd1";
      };
    };
    "optionator-0.8.3" = {
      name = "optionator";
      packageName = "optionator";
      version = "0.8.3";
      src = fetchurl {
        url = "https://registry.npmjs.org/optionator/-/optionator-0.8.3.tgz";
        sha512 =
          "+IW9pACdk3XWmmTXG8m3upGUJst5XRGzxMRjXzAuJ1XnIFNvfhjjIuYkDvysnPQ7qzqVzLt78BCruntqRhWQbA==";
      };
    };
    "pac-proxy-agent-5.0.0" = {
      name = "pac-proxy-agent";
      packageName = "pac-proxy-agent";
      version = "5.0.0";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/pac-proxy-agent/-/pac-proxy-agent-5.0.0.tgz";
        sha512 =
          "CcFG3ZtnxO8McDigozwE3AqAw15zDvGH+OjXO4kzf7IkEKkQ4gxQ+3sdF50WmhQ4P/bVusXcqNE2S3XrNURwzQ==";
      };
    };
    "pac-resolver-5.0.0" = {
      name = "pac-resolver";
      packageName = "pac-resolver";
      version = "5.0.0";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/pac-resolver/-/pac-resolver-5.0.0.tgz";
        sha512 =
          "H+/A6KitiHNNW+bxBKREk2MCGSxljfqRX76NjummWEYIat7ldVXRU3dhRIE3iXZ0nvGBk6smv3nntxKkzRL8NA==";
      };
    };
    "path-is-absolute-1.0.1" = {
      name = "path-is-absolute";
      packageName = "path-is-absolute";
      version = "1.0.1";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/path-is-absolute/-/path-is-absolute-1.0.1.tgz";
        sha1 = "174b9268735534ffbc7ace6bf53a5a9e1b5c5f5f";
      };
    };
    "path-parse-1.0.7" = {
      name = "path-parse";
      packageName = "path-parse";
      version = "1.0.7";
      src = fetchurl {
        url = "https://registry.npmjs.org/path-parse/-/path-parse-1.0.7.tgz";
        sha512 =
          "LDJzPVEEEPR+y48z93A0Ed0yXb8pAByGWo/k5YYdYgpY2/2EsOsksJrq7lOHxryrVOn1ejG6oAp8ahvOIQD8sw==";
      };
    };
    "picomatch-2.3.1" = {
      name = "picomatch";
      packageName = "picomatch";
      version = "2.3.1";
      src = fetchurl {
        url = "https://registry.npmjs.org/picomatch/-/picomatch-2.3.1.tgz";
        sha512 =
          "JU3teHTNjmE2VCGFzuY8EXzCDVwEqB2a8fsIvwaStHhAWJEeVd1o1QD80CU6+ZdEXXSLbSsuLwJjkCBWqRQUVA==";
      };
    };
    "pony-cause-1.1.1" = {
      name = "pony-cause";
      packageName = "pony-cause";
      version = "1.1.1";
      src = fetchurl {
        url = "https://registry.npmjs.org/pony-cause/-/pony-cause-1.1.1.tgz";
        sha512 =
          "PxkIc/2ZpLiEzQXu5YRDOUgBlfGYBY8156HY5ZcRAwwonMk5W/MrJP2LLkG/hF7GEQzaHo2aS7ho6ZLCOvf+6g==";
      };
    };
    "prelude-ls-1.1.2" = {
      name = "prelude-ls";
      packageName = "prelude-ls";
      version = "1.1.2";
      src = fetchurl {
        url = "https://registry.npmjs.org/prelude-ls/-/prelude-ls-1.1.2.tgz";
        sha1 = "21932a549f5e52ffd9a827f570e04be62a97da54";
      };
    };
    "printable-characters-1.0.42" = {
      name = "printable-characters";
      packageName = "printable-characters";
      version = "1.0.42";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/printable-characters/-/printable-characters-1.0.42.tgz";
        sha1 = "3f18e977a9bd8eb37fcc4ff5659d7be90868b3d8";
      };
    };
    "proxy-agent-5.0.0" = {
      name = "proxy-agent";
      packageName = "proxy-agent";
      version = "5.0.0";
      src = fetchurl {
        url = "https://registry.npmjs.org/proxy-agent/-/proxy-agent-5.0.0.tgz";
        sha512 =
          "gkH7BkvLVkSfX9Dk27W6TyNOWWZWRilRfk1XxGNWOYJ2TuedAv1yFpCaU9QSBmBe716XOTNpYNOzhysyw8xn7g==";
      };
    };
    "proxy-from-env-1.1.0" = {
      name = "proxy-from-env";
      packageName = "proxy-from-env";
      version = "1.1.0";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/proxy-from-env/-/proxy-from-env-1.1.0.tgz";
        sha512 =
          "D+zkORCbA9f1tdWRK0RaCR3GPv50cMxcrz4X8k5LTSUD1Dkw47mKJEZQNunItRTkWwgtaUSo1RVFRIG9ZXiFYg==";
      };
    };
    "punycode-2.1.1" = {
      name = "punycode";
      packageName = "punycode";
      version = "2.1.1";
      src = fetchurl {
        url = "https://registry.npmjs.org/punycode/-/punycode-2.1.1.tgz";
        sha512 =
          "XRsRjdf+j5ml+y/6GKHPZbrF/8p2Yga0JPtdqTIY2Xe5ohJPD9saDJJLPvp9+NSBprVvevdXZybnj2cv8OEd0A==";
      };
    };
    "queue-microtask-1.2.3" = {
      name = "queue-microtask";
      packageName = "queue-microtask";
      version = "1.2.3";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/queue-microtask/-/queue-microtask-1.2.3.tgz";
        sha512 =
          "NuaNSa6flKT5JaSYQzJok04JzTL1CA6aGhv5rfLW3PgqA+M2ChpZQnAC8h8i4ZFkBS8X5RqkDBHA7r4hej3K9A==";
      };
    };
    "raw-body-2.5.1" = {
      name = "raw-body";
      packageName = "raw-body";
      version = "2.5.1";
      src = fetchurl {
        url = "https://registry.npmjs.org/raw-body/-/raw-body-2.5.1.tgz";
        sha512 =
          "qqJBtEyVgS0ZmPGdCFPWJ3FreoqvG4MVQln/kCgF7Olq95IbOp0/BWyMwbdtn4VTvkM8Y7khCQ2Xgk/tcrCXig==";
      };
    };
    "readable-stream-1.1.14" = {
      name = "readable-stream";
      packageName = "readable-stream";
      version = "1.1.14";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/readable-stream/-/readable-stream-1.1.14.tgz";
        sha1 = "7cf4c54ef648e3813084c636dd2079e166c081d9";
      };
    };
    "require-directory-2.1.1" = {
      name = "require-directory";
      packageName = "require-directory";
      version = "2.1.1";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/require-directory/-/require-directory-2.1.1.tgz";
        sha1 = "8c64ad5fd30dab1c976e2344ffe7f792a6a6df42";
      };
    };
    "require-from-string-2.0.2" = {
      name = "require-from-string";
      packageName = "require-from-string";
      version = "2.0.2";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/require-from-string/-/require-from-string-2.0.2.tgz";
        sha512 =
          "Xf0nWe6RseziFMu+Ap9biiUbmplq6S9/p+7w7YXP/JBHhrUDDUhwa+vANyubuqfZWTveU//DYVGsDG7RKL/vEw==";
      };
    };
    "reserved-0.1.2" = {
      name = "reserved";
      packageName = "reserved";
      version = "0.1.2";
      src = fetchurl {
        url = "https://registry.npmjs.org/reserved/-/reserved-0.1.2.tgz";
        sha1 = "707b1246a3269f755da7cfcf9af6f4983bef105c";
      };
    };
    "resolve-1.22.0" = {
      name = "resolve";
      packageName = "resolve";
      version = "1.22.0";
      src = fetchurl {
        url = "https://registry.npmjs.org/resolve/-/resolve-1.22.0.tgz";
        sha512 =
          "Hhtrw0nLeSrFQ7phPp4OOcVjLPIeMnRlr5mcnVuMe7M/7eBn98A3hmFRLoFo3DLZkivSYwhRUJTyPyWAk56WLw==";
      };
    };
    "reusify-1.0.4" = {
      name = "reusify";
      packageName = "reusify";
      version = "1.0.4";
      src = fetchurl {
        url = "https://registry.npmjs.org/reusify/-/reusify-1.0.4.tgz";
        sha512 =
          "U9nH88a3fc/ekCF1l0/UP1IosiuIjyTh7hBvXVMHYgVcfGvt897Xguj2UOLDeI5BG2m7/uwyaLVT6fbtCwTyzw==";
      };
    };
    "rollup-2.67.3" = {
      name = "rollup";
      packageName = "rollup";
      version = "2.67.3";
      src = fetchurl {
        url = "https://registry.npmjs.org/rollup/-/rollup-2.67.3.tgz";
        sha512 =
          "G/x1vUwbGtP6O5ZM8/sWr8+p7YfZhI18pPqMRtMYMWSbHjKZ/ajHGiM+GWNTlWyOR0EHIdT8LHU+Z4ciIZ1oBw==";
      };
    };
    "run-parallel-1.2.0" = {
      name = "run-parallel";
      packageName = "run-parallel";
      version = "1.2.0";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/run-parallel/-/run-parallel-1.2.0.tgz";
        sha512 =
          "5l4VyZR86LZ/lDxZTR6jqL8AFE2S0IFLMP26AbjsLVADxHdhB/c0GUsH+y39UfCi3dzz8OlQuPmnaJOMoDHQBA==";
      };
    };
    "safe-stable-stringify-1.1.1" = {
      name = "safe-stable-stringify";
      packageName = "safe-stable-stringify";
      version = "1.1.1";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/safe-stable-stringify/-/safe-stable-stringify-1.1.1.tgz";
        sha512 =
          "ERq4hUjKDbJfE4+XtZLFPCDi8Vb1JqaxAPTxWFLBx8XcAlf9Bda/ZJdVezs/NAfsMQScyIlUMx+Yeu7P7rx5jw==";
      };
    };
    "safer-buffer-2.1.2" = {
      name = "safer-buffer";
      packageName = "safer-buffer";
      version = "2.1.2";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/safer-buffer/-/safer-buffer-2.1.2.tgz";
        sha512 =
          "YZo3K82SD7Riyi0E1EQPojLz7kpepnSQI9IyPbHHg1XXXevb5dJI7tpyN2ADxGcQbHG7vcyRHk0cbwqcQriUtg==";
      };
    };
    "setprototypeof-1.2.0" = {
      name = "setprototypeof";
      packageName = "setprototypeof";
      version = "1.2.0";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/setprototypeof/-/setprototypeof-1.2.0.tgz";
        sha512 =
          "E5LDX7Wrp85Kil5bhZv46j8jOeboKq5JMmYM3gVGdGH8xFpPWXUMsNrlODCrkoxMEeNi/XZIwuRvY4XNwYMJpw==";
      };
    };
    "simple-eval-1.0.0" = {
      name = "simple-eval";
      packageName = "simple-eval";
      version = "1.0.0";
      src = fetchurl {
        url = "https://registry.npmjs.org/simple-eval/-/simple-eval-1.0.0.tgz";
        sha512 =
          "kpKJR+bqTscgC0xuAl2xHN6bB12lHjC2DCUfqjAx19bQyO3R2EVLOurm3H9AUltv/uFVcSCVNc6faegR+8NYLw==";
      };
    };
    "smart-buffer-4.2.0" = {
      name = "smart-buffer";
      packageName = "smart-buffer";
      version = "4.2.0";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/smart-buffer/-/smart-buffer-4.2.0.tgz";
        sha512 =
          "94hK0Hh8rPqQl2xXc3HsaBoOXKV20MToPkcXvwbISWLEs+64sBq5kFgn2kJDHb1Pry9yrP0dxrCI9RRci7RXKg==";
      };
    };
    "socks-2.6.2" = {
      name = "socks";
      packageName = "socks";
      version = "2.6.2";
      src = fetchurl {
        url = "https://registry.npmjs.org/socks/-/socks-2.6.2.tgz";
        sha512 =
          "zDZhHhZRY9PxRruRMR7kMhnf3I8hDs4S3f9RecfnGxvcBHQcKcIH/oUcEWffsfl1XxdYlA7nnlGbbTvPz9D8gA==";
      };
    };
    "socks-proxy-agent-5.0.1" = {
      name = "socks-proxy-agent";
      packageName = "socks-proxy-agent";
      version = "5.0.1";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/socks-proxy-agent/-/socks-proxy-agent-5.0.1.tgz";
        sha512 =
          "vZdmnjb9a2Tz6WEQVIurybSwElwPxMZaIc7PzqbJTrezcKNznv6giT7J7tZDZ1BojVaa1jvO/UiUdhDVB0ACoQ==";
      };
    };
    "source-map-0.6.1" = {
      name = "source-map";
      packageName = "source-map";
      version = "0.6.1";
      src = fetchurl {
        url = "https://registry.npmjs.org/source-map/-/source-map-0.6.1.tgz";
        sha512 =
          "UjgapumWlbMhkBgzT7Ykc5YXUT46F0iKu8SGXq0bcwP5dz/h0Plj6enJqjz1Zbq2l5WaqYnrVbwWOWMyF3F47g==";
      };
    };
    "sourcemap-codec-1.4.8" = {
      name = "sourcemap-codec";
      packageName = "sourcemap-codec";
      version = "1.4.8";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/sourcemap-codec/-/sourcemap-codec-1.4.8.tgz";
        sha512 =
          "9NykojV5Uih4lgo5So5dtw+f0JgJX30KCNI8gwhz2J9A15wD0Ml6tjHKwf6fTSa6fAdVBdZeNOs9eJ71qCk8vA==";
      };
    };
    "stacktracey-2.1.8" = {
      name = "stacktracey";
      packageName = "stacktracey";
      version = "2.1.8";
      src = fetchurl {
        url = "https://registry.npmjs.org/stacktracey/-/stacktracey-2.1.8.tgz";
        sha512 =
          "Kpij9riA+UNg7TnphqjH7/CzctQ/owJGNbFkfEeve4Z4uxT5+JapVLFXcsurIfN34gnTWZNJ/f7NMG0E8JDzTw==";
      };
    };
    "statuses-2.0.1" = {
      name = "statuses";
      packageName = "statuses";
      version = "2.0.1";
      src = fetchurl {
        url = "https://registry.npmjs.org/statuses/-/statuses-2.0.1.tgz";
        sha512 =
          "RwNA9Z/7PrK06rYLIzFMlaF+l73iwpzsqRIFgbMLbTcLD6cOao82TaWefPXQvB2fOC4AjuYSEndS7N/mTCbkdQ==";
      };
    };
    "string-width-4.2.3" = {
      name = "string-width";
      packageName = "string-width";
      version = "4.2.3";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/string-width/-/string-width-4.2.3.tgz";
        sha512 =
          "wKyQRQpjJ0sIp62ErSZdGsjMJWsap5oRNihHhu6G7JVO/9jIB6UyevL+tXuOqrng8j/cxKTWyWUwvSTriiZz/g==";
      };
    };
    "string_decoder-0.10.31" = {
      name = "string_decoder";
      packageName = "string_decoder";
      version = "0.10.31";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/string_decoder/-/string_decoder-0.10.31.tgz";
        sha1 = "62e203bc41766c6c28c9fc84301dab1c5310fa94";
      };
    };
    "strip-ansi-6.0.1" = {
      name = "strip-ansi";
      packageName = "strip-ansi";
      version = "6.0.1";
      src = fetchurl {
        url = "https://registry.npmjs.org/strip-ansi/-/strip-ansi-6.0.1.tgz";
        sha512 =
          "Y38VPSHcqkFrCpFnQ9vuSXmquuv5oXOKpGeT6aGrr3o3Gc9AlVa6JBfUSOCnbxGGZF+/0ooI7KrPuUSztUdU5A==";
      };
    };
    "supports-color-7.2.0" = {
      name = "supports-color";
      packageName = "supports-color";
      version = "7.2.0";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/supports-color/-/supports-color-7.2.0.tgz";
        sha512 =
          "qpCAvRl9stuOHveKsn7HncJRvv501qIacKzQlO/+Lwxc9+0q2wLyv4Dfvt80/DPn2pqOBsJdDiogXGR9+OvwRw==";
      };
    };
    "supports-preserve-symlinks-flag-1.0.0" = {
      name = "supports-preserve-symlinks-flag";
      packageName = "supports-preserve-symlinks-flag";
      version = "1.0.0";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/supports-preserve-symlinks-flag/-/supports-preserve-symlinks-flag-1.0.0.tgz";
        sha512 =
          "ot0WnXS9fgdkgIcePe6RHNk1WA8+muPa6cSjeR3V8K27q9BB1rTE3R1p7Hv0z1ZyAc8s6Vvv8DIyWf681MAt0w==";
      };
    };
    "text-table-0.2.0" = {
      name = "text-table";
      packageName = "text-table";
      version = "0.2.0";
      src = fetchurl {
        url = "https://registry.npmjs.org/text-table/-/text-table-0.2.0.tgz";
        sha1 = "7f5ee823ae805207c00af2df4a84ec3fcfa570b4";
      };
    };
    "to-regex-range-5.0.1" = {
      name = "to-regex-range";
      packageName = "to-regex-range";
      version = "5.0.1";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/to-regex-range/-/to-regex-range-5.0.1.tgz";
        sha512 =
          "65P7iz6X5yEr1cwcgvQxbbIw7Uk3gOy5dIdtZ4rDveLqhrdJP+Li/Hx6tyK0NEb+2GCyneCMJiGqrADCSNk8sQ==";
      };
    };
    "toidentifier-1.0.1" = {
      name = "toidentifier";
      packageName = "toidentifier";
      version = "1.0.1";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/toidentifier/-/toidentifier-1.0.1.tgz";
        sha512 =
          "o5sSPKEkg/DIQNmH43V0/uerLrpzVedkUh8tGNvaeXpfpuwjKenlSox/2O/BTlZUtEe+JG7s5YhEz608PlAHRA==";
      };
    };
    "tr46-0.0.3" = {
      name = "tr46";
      packageName = "tr46";
      version = "0.0.3";
      src = fetchurl {
        url = "https://registry.npmjs.org/tr46/-/tr46-0.0.3.tgz";
        sha1 = "8184fd347dac9cdc185992f3a6622e14b9d9ab6a";
      };
    };
    "tslib-1.14.1" = {
      name = "tslib";
      packageName = "tslib";
      version = "1.14.1";
      src = fetchurl {
        url = "https://registry.npmjs.org/tslib/-/tslib-1.14.1.tgz";
        sha512 =
          "Xni35NKzjgMrwevysHTCArtLDpPvye8zV/0E4EyYn43P7/7qvQwPh9BGkHewbMulVntbigmcT7rdX3BNo9wRJg==";
      };
    };
    "tslib-2.4.0" = {
      name = "tslib";
      packageName = "tslib";
      version = "2.4.0";
      src = fetchurl {
        url = "https://registry.npmjs.org/tslib/-/tslib-2.4.0.tgz";
        sha512 =
          "d6xOpEDfsi2CZVlPQzGeux8XMwLT9hssAsaPYExaQMuYskwb+x1x7J371tWlbBdWHroy99KnVB6qIkUbs5X3UQ==";
      };
    };
    "type-check-0.3.2" = {
      name = "type-check";
      packageName = "type-check";
      version = "0.3.2";
      src = fetchurl {
        url = "https://registry.npmjs.org/type-check/-/type-check-0.3.2.tgz";
        sha1 = "5884cab512cf1d355e3fb784f30804b2b520db72";
      };
    };
    "universalify-0.1.2" = {
      name = "universalify";
      packageName = "universalify";
      version = "0.1.2";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/universalify/-/universalify-0.1.2.tgz";
        sha512 =
          "rBJeI5CXAlmy1pV+617WB9J63U6XcazHHF2f2dbJix4XzpUF0RS3Zbj0FGIOCAva5P/d/GBOYaACQ1w+0azUkg==";
      };
    };
    "unpipe-1.0.0" = {
      name = "unpipe";
      packageName = "unpipe";
      version = "1.0.0";
      src = fetchurl {
        url = "https://registry.npmjs.org/unpipe/-/unpipe-1.0.0.tgz";
        sha1 = "b2bf4ee8514aae6165b4817829d21b2ef49904ec";
      };
    };
    "uri-js-4.4.1" = {
      name = "uri-js";
      packageName = "uri-js";
      version = "4.4.1";
      src = fetchurl {
        url = "https://registry.npmjs.org/uri-js/-/uri-js-4.4.1.tgz";
        sha512 =
          "7rKUyy33Q1yc98pQ1DAmLtwX109F7TIfWlW1Ydo8Wl1ii1SeHieeh0HHfPeL2fMXK6z0s8ecKs9frCuLJvndBg==";
      };
    };
    "urijs-1.19.11" = {
      name = "urijs";
      packageName = "urijs";
      version = "1.19.11";
      src = fetchurl {
        url = "https://registry.npmjs.org/urijs/-/urijs-1.19.11.tgz";
        sha512 =
          "HXgFDgDommxn5/bIv0cnQZsPhHDA90NPHD6+c/v21U5+Sx5hoP8+dP9IZXBU1gIfvdRfhG8cel9QNPeionfcCQ==";
      };
    };
    "utility-types-3.10.0" = {
      name = "utility-types";
      packageName = "utility-types";
      version = "3.10.0";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/utility-types/-/utility-types-3.10.0.tgz";
        sha512 =
          "O11mqxmi7wMKCo6HKFt5AhO4BwY3VV68YU07tgxfz8zJTIxr4BpsezN49Ffwy9j3ZpwwJp4fkRwjRzq3uWE6Rg==";
      };
    };
    "validate-npm-package-name-3.0.0" = {
      name = "validate-npm-package-name";
      packageName = "validate-npm-package-name";
      version = "3.0.0";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/validate-npm-package-name/-/validate-npm-package-name-3.0.0.tgz";
        sha1 = "5fa912d81eb7d0c74afc140de7317f0ca7df437e";
      };
    };
    "vm2-3.9.9" = {
      name = "vm2";
      packageName = "vm2";
      version = "3.9.9";
      src = fetchurl {
        url = "https://registry.npmjs.org/vm2/-/vm2-3.9.9.tgz";
        sha512 =
          "xwTm7NLh/uOjARRBs8/95H0e8fT3Ukw5D/JJWhxMbhKzNh1Nu981jQKvkep9iKYNxzlVrdzD0mlBGkDKZWprlw==";
      };
    };
    "webidl-conversions-3.0.1" = {
      name = "webidl-conversions";
      packageName = "webidl-conversions";
      version = "3.0.1";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/webidl-conversions/-/webidl-conversions-3.0.1.tgz";
        sha1 = "24534275e2a7bc6be7bc86611cc16ae0a5654871";
      };
    };
    "whatwg-url-5.0.0" = {
      name = "whatwg-url";
      packageName = "whatwg-url";
      version = "5.0.0";
      src = fetchurl {
        url = "https://registry.npmjs.org/whatwg-url/-/whatwg-url-5.0.0.tgz";
        sha1 = "966454e8765462e37644d3626f6742ce8b70965d";
      };
    };
    "wolfy87-eventemitter-5.2.9" = {
      name = "wolfy87-eventemitter";
      packageName = "wolfy87-eventemitter";
      version = "5.2.9";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/wolfy87-eventemitter/-/wolfy87-eventemitter-5.2.9.tgz";
        sha512 =
          "P+6vtWyuDw+MB01X7UeF8TaHBvbCovf4HPEMF/SV7BdDc1SMTiBy13SRD71lQh4ExFTG1d/WNzDGDCyOKSMblw==";
      };
    };
    "word-wrap-1.2.3" = {
      name = "word-wrap";
      packageName = "word-wrap";
      version = "1.2.3";
      src = fetchurl {
        url = "https://registry.npmjs.org/word-wrap/-/word-wrap-1.2.3.tgz";
        sha512 =
          "Hz/mrNwitNRh/HUAtM/VT/5VH+ygD6DV7mYKZAtHOrbs8U7lvPS6xf7EJKMF0uW1KJCl0H701g3ZGus+muE5vQ==";
      };
    };
    "wrap-ansi-7.0.0" = {
      name = "wrap-ansi";
      packageName = "wrap-ansi";
      version = "7.0.0";
      src = fetchurl {
        url = "https://registry.npmjs.org/wrap-ansi/-/wrap-ansi-7.0.0.tgz";
        sha512 =
          "YVGIj2kamLSTxw6NsZjoBxfSwsn0ycdesmc4p+Q21c5zPuZ1pl+NfxVdxPtdHvmNVOQ6XSYG4AUtyt/Fi7D16Q==";
      };
    };
    "wrappy-1.0.2" = {
      name = "wrappy";
      packageName = "wrappy";
      version = "1.0.2";
      src = fetchurl {
        url = "https://registry.npmjs.org/wrappy/-/wrappy-1.0.2.tgz";
        sha1 = "b5243d8f3ec1aa35f1364605bc0d1036e30ab69f";
      };
    };
    "xregexp-2.0.0" = {
      name = "xregexp";
      packageName = "xregexp";
      version = "2.0.0";
      src = fetchurl {
        url = "https://registry.npmjs.org/xregexp/-/xregexp-2.0.0.tgz";
        sha1 = "52a63e56ca0b84a7f3a5f3d61872f126ad7a5943";
      };
    };
    "y18n-5.0.8" = {
      name = "y18n";
      packageName = "y18n";
      version = "5.0.8";
      src = fetchurl {
        url = "https://registry.npmjs.org/y18n/-/y18n-5.0.8.tgz";
        sha512 =
          "0pfFzegeDWJHJIAmTLRP2DwHjdF5s7jo9tuztdQxAhINCdvS+3nGINqPd00AphqJR/0LhANUS6/+7SCb98YOfA==";
      };
    };
    "yallist-3.1.1" = {
      name = "yallist";
      packageName = "yallist";
      version = "3.1.1";
      src = fetchurl {
        url = "https://registry.npmjs.org/yallist/-/yallist-3.1.1.tgz";
        sha512 =
          "a4UGQaWPH59mOXUYnAG2ewncQS4i4F43Tv3JoAM+s2VDAmS9NsK8GpDMLrCHPksFT7h3K6TOoUNn2pb7RoXx4g==";
      };
    };
    "yargs-17.3.1" = {
      name = "yargs";
      packageName = "yargs";
      version = "17.3.1";
      src = fetchurl {
        url = "https://registry.npmjs.org/yargs/-/yargs-17.3.1.tgz";
        sha512 =
          "WUANQeVgjLbNsEmGk20f+nlHgOqzRFpiGWVaBrYGYIGANIIu3lWjoyi0fNlFmJkvfhCZ6BXINe7/W2O2bV4iaA==";
      };
    };
    "yargs-parser-21.0.1" = {
      name = "yargs-parser";
      packageName = "yargs-parser";
      version = "21.0.1";
      src = fetchurl {
        url =
          "https://registry.npmjs.org/yargs-parser/-/yargs-parser-21.0.1.tgz";
        sha512 =
          "9BK1jFpLzJROCI5TzwZL/TU4gqjK5xiHV/RfWLOahrjAko/e4DJkRDZQXfvqAsiZzzYhgAzbgz6lg48jcm4GLg==";
      };
    };
  };
in buildNodePackage {
  name = "_at_stoplight_slash_spectral-cli";
  packageName = "@stoplight/spectral-cli";
  version = "6.3.0";
  src = fetchurl {
    url =
      "https://registry.npmjs.org/@stoplight/spectral-cli/-/spectral-cli-6.3.0.tgz";
    sha512 =
      "AP9RUMHJGzT6tMd3KWyweTbNinslcNWfnwg2r47Bw1NRffsZ2/CPLwW4L/Ay7yAdNDV5CaNsvF3OjhEqKk5jrQ==";
  };
  dependencies = [
    sources."@asyncapi/specs-2.14.0"
    sources."@jsep-plugin/regex-1.0.2"
    sources."@jsep-plugin/ternary-1.1.2"
    sources."@nodelib/fs.scandir-2.1.5"
    sources."@nodelib/fs.stat-2.0.5"
    sources."@nodelib/fs.walk-1.2.8"
    sources."@rollup/plugin-commonjs-20.0.0"
    (sources."@rollup/pluginutils-3.1.0" // {
      dependencies = [ sources."estree-walker-1.0.1" ];
    })
    sources."@stoplight/better-ajv-errors-1.0.1"
    sources."@stoplight/json-3.17.0"
    (sources."@stoplight/json-ref-readers-1.2.2" // {
      dependencies = [ sources."tslib-1.14.1" ];
    })
    sources."@stoplight/json-ref-resolver-3.1.3"
    sources."@stoplight/lifecycle-2.3.2"
    sources."@stoplight/ordered-object-literal-1.0.2"
    sources."@stoplight/path-1.3.2"
    (sources."@stoplight/spectral-core-1.12.1" // {
      dependencies = [
        (sources."@stoplight/json-3.17.2" // {
          dependencies = [ sources."@stoplight/types-12.5.0" ];
        })
        sources."@stoplight/types-13.0.0"
        sources."minimatch-3.0.4"
      ];
    })
    sources."@stoplight/spectral-formats-1.2.0"
    (sources."@stoplight/spectral-functions-1.6.1" // {
      dependencies = [ sources."@stoplight/json-3.17.2" ];
    })
    sources."@stoplight/spectral-parsers-1.0.1"
    sources."@stoplight/spectral-ref-resolver-1.0.1"
    (sources."@stoplight/spectral-ruleset-bundler-1.2.1" // {
      dependencies = [ sources."@rollup/plugin-commonjs-21.1.0" ];
    })
    sources."@stoplight/spectral-ruleset-migrator-1.7.3"
    (sources."@stoplight/spectral-rulesets-1.8.0" // {
      dependencies = [ sources."@stoplight/types-12.5.0" ];
    })
    sources."@stoplight/spectral-runtime-1.1.2"
    sources."@stoplight/types-12.3.0"
    sources."@stoplight/yaml-4.2.2"
    sources."@stoplight/yaml-ast-parser-0.0.48"
    sources."@tootallnate/once-1.1.2"
    sources."@types/estree-0.0.39"
    sources."@types/json-schema-7.0.11"
    sources."@types/node-17.0.31"
    sources."@types/urijs-1.19.19"
    sources."abort-controller-3.0.0"
    sources."acorn-8.7.1"
    sources."acorn-walk-8.2.0"
    sources."agent-base-6.0.2"
    sources."ajv-8.11.0"
    sources."ajv-draft-04-1.0.0"
    sources."ajv-errors-3.0.0"
    sources."ajv-formats-2.1.1"
    sources."ansi-regex-5.0.1"
    sources."ansi-styles-4.3.0"
    sources."as-table-1.0.55"
    sources."ast-types-0.14.2"
    sources."astring-1.8.3"
    sources."balanced-match-1.0.2"
    sources."blueimp-md5-2.18.0"
    sources."brace-expansion-1.1.11"
    sources."braces-3.0.2"
    sources."builtins-1.0.3"
    sources."bytes-3.1.2"
    sources."chalk-4.1.2"
    sources."cliui-7.0.4"
    sources."color-convert-2.0.1"
    sources."color-name-1.1.4"
    sources."commondir-1.0.1"
    sources."concat-map-0.0.1"
    sources."core-util-is-1.0.3"
    sources."data-uri-to-buffer-3.0.1"
    sources."debug-4.3.4"
    sources."deep-is-0.1.4"
    (sources."degenerator-3.0.2" // {
      dependencies = [ sources."ast-types-0.13.4" ];
    })
    sources."depd-2.0.0"
    sources."dependency-graph-0.11.0"
    sources."emoji-regex-8.0.0"
    sources."eol-0.9.1"
    sources."escalade-3.1.1"
    sources."escodegen-1.14.3"
    sources."esprima-4.0.1"
    sources."estraverse-4.3.0"
    sources."estree-walker-2.0.2"
    sources."esutils-2.0.3"
    sources."event-target-shim-5.0.1"
    sources."fast-deep-equal-3.1.3"
    sources."fast-glob-3.2.7"
    sources."fast-levenshtein-2.0.6"
    sources."fast-memoize-2.5.2"
    sources."fastq-1.13.0"
    sources."file-uri-to-path-2.0.0"
    sources."fill-range-7.0.1"
    sources."fs-extra-8.1.0"
    sources."fs.realpath-1.0.0"
    sources."fsevents-2.3.2"
    sources."ftp-0.3.10"
    sources."function-bind-1.1.1"
    sources."get-caller-file-2.0.5"
    (sources."get-source-2.0.12" // {
      dependencies = [ sources."data-uri-to-buffer-2.0.2" ];
    })
    sources."get-uri-3.0.2"
    sources."glob-7.2.0"
    sources."glob-parent-5.1.2"
    sources."graceful-fs-4.2.10"
    sources."has-1.0.3"
    sources."has-flag-4.0.0"
    sources."http-errors-2.0.0"
    sources."http-proxy-agent-4.0.1"
    sources."https-proxy-agent-5.0.1"
    sources."iconv-lite-0.4.24"
    sources."immer-9.0.12"
    sources."inflight-1.0.6"
    sources."inherits-2.0.4"
    sources."ip-1.1.5"
    sources."is-core-module-2.9.0"
    sources."is-extglob-2.1.1"
    sources."is-fullwidth-code-point-3.0.0"
    sources."is-glob-4.0.3"
    sources."is-number-7.0.0"
    sources."is-reference-1.2.1"
    sources."isarray-0.0.1"
    sources."jsep-1.3.6"
    sources."json-schema-traverse-1.0.0"
    sources."jsonc-parser-2.2.1"
    sources."jsonfile-4.0.0"
    sources."jsonpath-plus-6.0.1"
    sources."jsonpointer-5.0.0"
    sources."leven-3.1.0"
    sources."levn-0.3.0"
    sources."lodash-4.17.21"
    sources."lodash.get-4.4.2"
    sources."lodash.set-4.3.2"
    sources."lodash.topath-4.5.2"
    sources."lru-cache-5.1.1"
    sources."magic-string-0.25.9"
    sources."merge2-1.4.1"
    sources."micromatch-4.0.5"
    sources."minimatch-3.1.2"
    sources."ms-2.1.2"
    sources."netmask-2.0.2"
    sources."nimma-0.2.0"
    sources."node-fetch-2.6.7"
    sources."once-1.4.0"
    sources."optionator-0.8.3"
    sources."pac-proxy-agent-5.0.0"
    sources."pac-resolver-5.0.0"
    sources."path-is-absolute-1.0.1"
    sources."path-parse-1.0.7"
    sources."picomatch-2.3.1"
    sources."pony-cause-1.1.1"
    sources."prelude-ls-1.1.2"
    sources."printable-characters-1.0.42"
    sources."proxy-agent-5.0.0"
    sources."proxy-from-env-1.1.0"
    sources."punycode-2.1.1"
    sources."queue-microtask-1.2.3"
    sources."raw-body-2.5.1"
    sources."readable-stream-1.1.14"
    sources."require-directory-2.1.1"
    sources."require-from-string-2.0.2"
    sources."reserved-0.1.2"
    sources."resolve-1.22.0"
    sources."reusify-1.0.4"
    sources."rollup-2.67.3"
    sources."run-parallel-1.2.0"
    sources."safe-stable-stringify-1.1.1"
    sources."safer-buffer-2.1.2"
    sources."setprototypeof-1.2.0"
    sources."simple-eval-1.0.0"
    sources."smart-buffer-4.2.0"
    sources."socks-2.6.2"
    sources."socks-proxy-agent-5.0.1"
    sources."source-map-0.6.1"
    sources."sourcemap-codec-1.4.8"
    sources."stacktracey-2.1.8"
    sources."statuses-2.0.1"
    sources."string-width-4.2.3"
    sources."string_decoder-0.10.31"
    sources."strip-ansi-6.0.1"
    sources."supports-color-7.2.0"
    sources."supports-preserve-symlinks-flag-1.0.0"
    sources."text-table-0.2.0"
    sources."to-regex-range-5.0.1"
    sources."toidentifier-1.0.1"
    sources."tr46-0.0.3"
    sources."tslib-2.4.0"
    sources."type-check-0.3.2"
    sources."universalify-0.1.2"
    sources."unpipe-1.0.0"
    sources."uri-js-4.4.1"
    sources."urijs-1.19.11"
    sources."utility-types-3.10.0"
    sources."validate-npm-package-name-3.0.0"
    sources."vm2-3.9.9"
    sources."webidl-conversions-3.0.1"
    sources."whatwg-url-5.0.0"
    sources."wolfy87-eventemitter-5.2.9"
    sources."word-wrap-1.2.3"
    sources."wrap-ansi-7.0.0"
    sources."wrappy-1.0.2"
    sources."xregexp-2.0.0"
    sources."y18n-5.0.8"
    sources."yallist-3.1.1"
    sources."yargs-17.3.1"
    sources."yargs-parser-21.0.1"
  ];
  buildInputs = globalBuildInputs;
  meta = {
    homepage = "https://github.com/stoplightio/spectral";
    license = "Apache-2.0";
  };
  npmFlags = "--only=production";
  production = true;
  bypassCache = true;
  reconstructLock = true;
}

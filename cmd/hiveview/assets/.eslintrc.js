module.exports = {
    "ignorePatterns": [
        ".eslintrc.js",
        "extlib/*"
    ],
    "extends": "eslint:recommended",
    "overrides": [],
    "env": {
        "browser": true,
        "es2021": true
    },
    "parserOptions": {
        "ecmaVersion": "latest",
        "sourceType": "module"
    },
    "rules": {
        "indent": [
            "error", 4
        ],
        "linebreak-style": [
            "error", "unix"
        ],
        "quotes": [
            "error", "single"
        ],
        "semi": [
            "error", "always"
        ],
        "no-unused-vars": [
            "error", { "vars": "all", "args": "none", "ignoreRestSiblings": false }
        ],
        "no-unused-private-class-members": [
            "error"
        ],
        "no-else-return": [
            "error"
        ],
        "no-unreachable-loop": [
            "error"
        ],
        "no-duplicate-imports": [
            "error"
        ],
        "no-unmodified-loop-condition": [
            "warn"
        ],
        // "no-use-before-define": [
        //     "warn"
        // ],
    }
}

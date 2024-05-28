// STEP 1: POLYFILL THE CRYPTO LIBRARY
const fs = require('fs');
const crypto = require("crypto").webcrypto;
const redis = require("redis")

globalThis.crypto = crypto;

// STEP 2: CREATE THE GO IMPORT OBJECT
require('./wasm_exec.js');

// STEP 3: IMPORT THE MIDDLEWARE BINARY
const wasmBin = require("./middleware.json");

// STEP 4: DECLARE NECESSARY UTILITY FUNCTIONS
function decode(encoded) {
    var binaryString =  Buffer.from(encoded, 'base64').toString('binary');
    var bytes = new Uint8Array(binaryString.length);
    for (var i = 0; i < binaryString.length; i++) {
        bytes[i] = binaryString.charCodeAt(i);
    }
    return bytes.buffer;
}

// STEP 5: IMPORT
const go = new Go();
const importObject = go.importObject;
WebAssembly.instantiate(decode(wasmBin), importObject).then(async (results) => {
    const instance = results.instance
    go.run(instance);
    console.log("WASM is Loaded")
}).catch((err)=>{
    console.log("Error running loadWASM script: ", err)
});

// STEP 6: EXPORT
// module.exports = function Layer8(req, res, next) { 
//     WASMMiddleware(req, res, next);
// };

class RedisClient {
    constructor(host, port, password) {
        this._client = 2;
        this.host = host;
        this.port = port;
        this.password = password;
    }
}

module.exports = {
    tunnel: (req, res, next) => {
        WASMMiddleware(req, res, next, null);
    },
    storage: (client) => {
        // currently only Redis is supported
        if (!(client instanceof RedisClient)) {
            throw new Error("Invalid Redis client instance");
        }
        return async (req, res, next) => {
            var rediscl = redis.createClient({
                url: `redis://${client.host}:${client.port}/${client.db || 0}`
            })
            rediscl.on('error', err => console.log('Redis Client Error', err));
            await rediscl.connect();

            StorageOptions(next, {
                _client: 2,
                host: client.host,
                port: client.port,
                password: client.password,
                db: client.db || 0,
                client: rediscl
            });
        }
    },
    static: (dir) => {
        return (req, res, next) => {
            ServeStatic(req, res, dir, fs);
        }
    },
    multipart: (options) => {
        return {
            single: (name) => {
                return (req, res, next) => {
                    const multi = ProcessMultipart(options, fs)
                    multi.single(req, res, next, name)
                }
            },
            array: (name) => {
                return (req, res, next) => {
                    const multi = ProcessMultipart(options, fs)
                    multi.array(req, res, next, name)
                }
            }
        }
    },
    RedisClient
}

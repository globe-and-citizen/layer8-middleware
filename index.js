// STEP 1: POLYFILL THE CRYPTO LIBRARY
const fs = require('fs');
const crypto = require("crypto").webcrypto;
globalThis.crypto = crypto;

// STEP 2: CREATE THE GO IMPORT OBJECT
require('./dist/wasm_exec.js');

// STEP 3: IMPORT THE MIDDLEWARE BINARY
const wasmBin = require("./dist/middleware.json");

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

module.exports = {
    tunnel: (req, res, next) => {
        WASMMiddleware(req, res, next);
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
    }
}
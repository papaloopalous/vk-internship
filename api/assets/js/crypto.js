function normalizeResponse(resp) {
  return {
    success:    resp.success,
    statusCode: resp.code,
    message:    resp.message,
    data:       resp.data
  };
}

class KeyExchange {
  constructor(prime, generator) {
    this.p = BigInt('0x' + prime);
    this.g = BigInt(generator);
    this.priv = BigInt('0x9876543210FEDCBA9876543210FEDCBA98765432');
    this.pub = this.modPow(this.g, this.priv, this.p);
  }

  modPow(b, e, m) {
    let r = 1n;
    b %= m;
    while (e) { if (e & 1n) r = (r * b) % m; b = (b * b) % m; e >>= 1n; }
    return r;
  }

  shared(serverPub) { 
    return this.modPow(BigInt(serverPub), this.priv, this.p); 
  }
}

let currentKeyExchange = null;

async function initializeKeyExchange() {
  if (currentKeyExchange) {
    console.log('Key exchange already in progress, waiting...');
    return currentKeyExchange;
  }

  console.log('Starting new key exchange');
  
  try {
    currentKeyExchange = (async () => {
      const paramsResponse = await fetch('/api/crypto-params', {
        method: 'GET',
        headers: { 'Content-Type': 'application/json' }
      });
      console.log('Got crypto params');
      
      const { success, data, message } = await paramsResponse.json();
      if (!success) throw new Error(message);

      const kex = new KeyExchange(data.prime, data.generator);
      
      const keyExchange = await fetch('/api/key-exchange', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ clientPublic: kex.pub.toString() })
      });
      console.log('Completed key exchange');

      const result = await keyExchange.json();
      if (!result.success) throw new Error(result.message);

      const sharedKeyHex = CryptoJS.SHA256(
        kex.shared(result.data.serverPublic).toString()
      ).toString(CryptoJS.enc.Hex);

      if (sharedKeyHex.length !== 64)
        throw new Error('Shared key must be 32 bytes (64 hex chars)');

      return sharedKeyHex;
    })();

    return await currentKeyExchange;
  } finally {
    currentKeyExchange = null;
  }
}

function deriveKeyAndIV(keyHexWA, saltWA) {
  let acc = CryptoJS.lib.WordArray.create();
  let prev = keyHexWA;

  while (acc.sigBytes < 48) {
    prev = CryptoJS.MD5(prev.concat(saltWA));
    acc = acc.concat(prev);
  }
  return {
    key: CryptoJS.lib.WordArray.create(acc.words.slice(0, 8)),
    iv: CryptoJS.lib.WordArray.create(acc.words.slice(8, 12))
  };
}

async function encryptData(plain) {
  if (typeof plain !== 'string')
    throw new TypeError('encryptData expects a string');

  const keyHex = await initializeKeyExchange();
  const salt = CryptoJS.lib.WordArray.random(8);
  const { key, iv } = deriveKeyAndIV(CryptoJS.enc.Hex.parse(keyHex), salt);

  const cipher = CryptoJS.AES.encrypt(
    CryptoJS.enc.Utf8.parse(plain),
    key,
    { iv, mode: CryptoJS.mode.CBC, padding: CryptoJS.pad.Pkcs7 }
  );

  const salted = CryptoJS.enc.Utf8.parse('Salted__').concat(salt)
                 .concat(cipher.ciphertext);
  return CryptoJS.enc.Base64.stringify(salted);
}
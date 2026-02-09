// SSL Pinning Bypass Script
Java.perform(function () {
    console.log("[*] SSL Pinning Bypass loaded");

    // TrustManager bypass
    var TrustManager = Java.registerClass({
        name: 'com.godroid.TrustManager',
        implements: [Java.use('javax.net.ssl.X509TrustManager')],
        methods: {
            checkClientTrusted: function (chain, authType) { },
            checkServerTrusted: function (chain, authType) { },
            getAcceptedIssuers: function () { return []; }
        }
    });

    // SSLContext bypass
    var SSLContext = Java.use('javax.net.ssl.SSLContext');
    SSLContext.init.overload('[Ljavax.net.ssl.KeyManager;', '[Ljavax.net.ssl.TrustManager;', 'java.security.SecureRandom').implementation = function (km, tm, sr) {
        console.log("[+] SSLContext.init() bypassed");
        this.init(km, [TrustManager.$new()], sr);
    };

    // OkHttp CertificatePinner bypass
    try {
        var CertificatePinner = Java.use('okhttp3.CertificatePinner');
        CertificatePinner.check.overload('java.lang.String', 'java.util.List').implementation = function (hostname, peerCertificates) {
            console.log("[+] OkHttp CertificatePinner bypassed for: " + hostname);
            return;
        };
        CertificatePinner.check.overload('java.lang.String', '[Ljava.security.cert.Certificate;').implementation = function (hostname, peerCertificates) {
            console.log("[+] OkHttp CertificatePinner bypassed for: " + hostname);
            return;
        };
    } catch (e) {
        console.log("[-] OkHttp not found or bypass failed: " + e);
    }

    // TrustManagerImpl bypass (Android)
    try {
        var TrustManagerImpl = Java.use('com.android.org.conscrypt.TrustManagerImpl');
        TrustManagerImpl.verifyChain.implementation = function (untrustedChain, trustAnchorChain, host, clientAuth, ocspData, tlsSctData) {
            console.log("[+] TrustManagerImpl.verifyChain bypassed for: " + host);
            return untrustedChain;
        };
    } catch (e) {
        console.log("[-] TrustManagerImpl bypass failed: " + e);
    }

    console.log("[+] SSL Pinning bypass active");
});

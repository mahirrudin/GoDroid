// PintooR - Combined SSL Pinning + Root Detection Bypass
Java.perform(function () {
    console.log("[*] PintooR - Combined Bypass loaded");
    console.log("[*] Bypassing SSL Pinning and Root Detection...");

    // ===== SSL PINNING BYPASS =====

    // TrustManager bypass
    try {
        var TrustManager = Java.registerClass({
            name: 'com.godroid.PintoorTrustManager',
            implements: [Java.use('javax.net.ssl.X509TrustManager')],
            methods: {
                checkClientTrusted: function (chain, authType) { },
                checkServerTrusted: function (chain, authType) { },
                getAcceptedIssuers: function () { return []; }
            }
        });

        var SSLContext = Java.use('javax.net.ssl.SSLContext');
        SSLContext.init.overload('[Ljavax.net.ssl.KeyManager;', '[Ljavax.net.ssl.TrustManager;', 'java.security.SecureRandom').implementation = function (km, tm, sr) {
            console.log("[+] SSL: SSLContext.init bypassed");
            this.init(km, [TrustManager.$new()], sr);
        };
    } catch (e) {
        console.log("[-] SSLContext bypass error: " + e);
    }

    // OkHttp bypass
    try {
        var CertificatePinner = Java.use('okhttp3.CertificatePinner');
        CertificatePinner.check.overload('java.lang.String', 'java.util.List').implementation = function (hostname, peerCertificates) {
            console.log("[+] SSL: OkHttp bypassed for " + hostname);
            return;
        };
    } catch (e) {
        // OkHttp might not be present
    }

    // ===== ROOT DETECTION BYPASS =====

    // File.exists bypass
    try {
        var File = Java.use('java.io.File');
        var originalExists = File.exists;
        File.exists.implementation = function () {
            var path = this.getAbsolutePath();
            var rootIndicators = ['su', 'magisk', 'superuser', 'busybox', 'xposed'];
            for (var i = 0; i < rootIndicators.length; i++) {
                if (path.toLowerCase().indexOf(rootIndicators[i]) !== -1) {
                    console.log("[+] ROOT: Hidden path " + path);
                    return false;
                }
            }
            return originalExists.call(this);
        };
    } catch (e) {
        console.log("[-] File.exists bypass error: " + e);
    }

    // Build.TAGS bypass
    try {
        var Build = Java.use('android.os.Build');
        Build.TAGS.value = "release-keys";
        console.log("[+] ROOT: Build.TAGS set to release-keys");
    } catch (e) {
        console.log("[-] Build.TAGS bypass error: " + e);
    }

    // Emulator detection bypass
    try {
        var Build = Java.use('android.os.Build');
        Build.HARDWARE.value = "qcom";
        Build.PRODUCT.value = "redfin";
        Build.MANUFACTURER.value = "Google";
        Build.BRAND.value = "google";
        Build.MODEL.value = "Pixel 5";
        Build.FINGERPRINT.value = "google/redfin/redfin:11/RQ3A.210805.001.A1/7474174:user/release-keys";
        console.log("[+] EMULATOR: Build properties spoofed");
    } catch (e) {
        console.log("[-] Build spoof error: " + e);
    }

    // Package manager bypass for root apps
    try {
        var PackageManager = Java.use('android.app.ApplicationPackageManager');
        PackageManager.getPackageInfo.overload('java.lang.String', 'int').implementation = function (packageName, flags) {
            var rootPackages = ['com.topjohnwu.magisk', 'eu.chainfire.supersu', 'com.koushikdutta.superuser'];
            for (var i = 0; i < rootPackages.length; i++) {
                if (packageName === rootPackages[i]) {
                    console.log("[+] ROOT: Hidden package " + packageName);
                    throw Java.use('android.content.pm.PackageManager$NameNotFoundException').$new();
                }
            }
            return this.getPackageInfo(packageName, flags);
        };
    } catch (e) {
        console.log("[-] PackageManager bypass error: " + e);
    }

    console.log("[+] PintooR bypass active - SSL Pinning and Root Detection disabled");
});

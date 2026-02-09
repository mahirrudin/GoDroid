// Root Detection Bypass Script
Java.perform(function () {
    console.log("[*] Root Detection Bypass loaded");

    // File.exists bypass for root paths
    var File = Java.use('java.io.File');
    File.exists.implementation = function () {
        var path = this.getAbsolutePath();
        var rootPaths = [
            'su', 'Superuser', 'superuser', 'magisk', 'Magisk',
            'busybox', 'xposed', 'titanium', 'substrate',
            '/system/app/Superuser', '/system/bin/su', '/system/xbin/su',
            '/sbin/su', '/data/local/xbin/su', '/data/local/bin/su',
            '/data/local/su', '/system/sd/xbin/su'
        ];
        for (var i = 0; i < rootPaths.length; i++) {
            if (path.indexOf(rootPaths[i]) !== -1) {
                console.log("[+] Blocked root path check: " + path);
                return false;
            }
        }
        return this.exists();
    };

    // Runtime.exec bypass
    var Runtime = Java.use('java.lang.Runtime');
    Runtime.exec.overload('java.lang.String').implementation = function (cmd) {
        if (cmd.indexOf('su') !== -1 || cmd.indexOf('which') !== -1) {
            console.log("[+] Blocked Runtime.exec: " + cmd);
            throw new Error("Command not found");
        }
        return this.exec(cmd);
    };

    // Build properties bypass
    var Build = Java.use('android.os.Build');
    Build.TAGS.value = "release-keys";
    Build.FINGERPRINT.value = Build.FINGERPRINT.value.replace("test-keys", "release-keys");

    // System property bypass
    try {
        var SystemProperties = Java.use('android.os.SystemProperties');
        SystemProperties.get.overload('java.lang.String').implementation = function (key) {
            if (key === 'ro.build.selinux') {
                return '1';
            }
            return this.get(key);
        };
    } catch (e) {
        console.log("[-] SystemProperties bypass failed: " + e);
    }

    // RootBeer bypass
    try {
        var RootBeer = Java.use('com.scottyab.rootbeer.RootBeer');
        RootBeer.isRooted.implementation = function () {
            console.log("[+] RootBeer.isRooted bypassed");
            return false;
        };
        RootBeer.isRootedWithoutBusyBoxCheck.implementation = function () {
            console.log("[+] RootBeer.isRootedWithoutBusyBoxCheck bypassed");
            return false;
        };
    } catch (e) {
        console.log("[-] RootBeer not found");
    }

    console.log("[+] Root detection bypass active");
});

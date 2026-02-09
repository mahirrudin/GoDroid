// Package apk provides APK analysis functionality.
package apk

// FrameworkPatterns maps framework names to their detection file patterns.
var FrameworkPatterns = map[string][]string{
	"Flutter": {
		"lib/*/libflutter.so",
		"lib/*/libapp.so",
		"assets/flutter_assets/",
	},
	"React Native": {
		"assets/index.android.bundle",
		"lib/*/libreactnativejni.so",
	},
	"Ionic + Cordova": {
		"assets/www/",
		"res/xml/config.xml",
	},
	"Ionic + Capacitor": {
		"assets/capacitor.config.json",
		"assets/capacitor.plugins.json",
	},
	"Framework7": {
		"assets/www/framework7.js",
		"assets/www/framework7.css",
		"assets/www/framework7-bundle.js",
	},
	"NativeScript": {
		"assets/app/package.json",
		"lib/*/libNativeScript.so",
	},
	"Xamarin": {
		"assemblies/",
		"lib/*/libmonodroid.so",
		"lib/*/libmonosgen-2.0.so",
	},
	"Unity": {
		"lib/*/libunity.so",
		"lib/*/libil2cpp.so",
		"assets/bin/Data/",
	},
	"Kotlin Multiplatform": {
		"kotlin/",
	},
	"Qt": {
		"lib/*/libQt5Core.so",
		"lib/*/libQt6Core.so",
	},
	"Cordova": {
		"assets/www/cordova.js",
		"assets/www/cordova_plugins.js",
	},
}

// SSLPinningPatterns maps SSL pinning types to smali patterns.
var SSLPinningPatterns = map[string][]string{
	"TrustManager SSL Pinning": {
		"javax/net/ssl/X509TrustManager;->checkServerTrusted",
		"javax/net/ssl/X509TrustManager;->getAcceptedIssuers",
		"javax/net/ssl/HostnameVerifier;->verify",
		"javax/net/ssl/HttpsURLConnection;->setDefaultHostnameVerifier",
		"javax/net/ssl/X509TrustManager;->checkClientTrusted",
	},
	"OkHttp3 Certificate Pinning": {
		"com/squareup/okhttp/CertificatePinner;",
		"okhttp3/CertificatePinner;",
		"okhttp3/CertificatePinner;->check",
		"okhttp3/CertificatePinner;->check$okhttp",
		"okhttp3/OkHttpClient$Builder;->certificatePinner",
	},
	"SSLSocketFactory": {
		"javax/net/ssl/HttpsURLConnection;->setSSLSocketFactory",
		"SSLSocketFactory.getDefault()",
		"javax/net/ssl/HttpsURLConnection",
	},
	"Trustkit Certificate Pinning": {
		"com/datatheorem/android/trustkit/pinning/OkHostnameVerifier;->verify",
		"com/datatheorem/android/trustkit/pinning/PinningTrustManager;->checkServerTrusted",
	},
	"Conscrypt TrustManagerImpl Pinning": {
		"com/android/org/conscrypt/TrustManagerImpl;->checkTrustedRecursive",
		"com/android/org/conscrypt/TrustManagerImpl;->verifyChain",
	},
	"Appcelerator Certificate Pinning": {
		"appcelerator/https/PinningTrustManager;->checkServerTrusted",
	},
	"Fabric Certificate Pinning": {
		"io/fabric/sdk/android/services/network/PinningTrustManager;->checkServerTrusted",
	},
	"Conscrypt OpenSSLSocketImpl Pinning": {
		"com/android/org/conscrypt/OpenSSLSocketImpl;->verifyCertificateChain",
	},
	"Conscrypt OpenSSLEngineSocketImpl Pinning": {
		"com/android/org/conscrypt/OpenSSLEngineSocketImpl;->verifyCertificateChain",
	},
	"Apache Harmony OpenSSLSocketImpl Pinning": {
		"org/apache/harmony/xnet/provider/jsse/OpenSSLSocketImpl;->verifyCertificateChain",
	},
	"PhoneGap Certificate Checker": {
		"nl/xservices/plugins/sslCertificateChecker;->execute",
	},
	"IBM MobileFirst Certificate Pinning": {
		"com/worklight/wlclient/api/WLClient;->pinTrustedCertificatePublicKey",
	},
	"IBM WorkLight HostNameVerifier Pinning": {
		"com/worklight/wlclient/certificatepinning/HostNameVerifierWithCertificatePinning;->verify",
	},
	"Conscrypt CertPinManager Pinning": {
		"com/android/org/conscrypt/CertPinManager;->checkChainPinning",
		"com/android/org/conscrypt/CertPinManager;->isChainValid",
	},
	"CWAC-Netsecurity CertPinManager Pinning": {
		"com/commonsware/cwac/netsecurity/conscrypt/CertPinManager;->isChainValid",
	},
	"Worklight Androidgap Pinning Plugin": {
		"com/worklight/androidgap/plugin/WLCertificatePinningPlugin;->execute",
	},
	"Netty Fingerprint TrustManager": {
		"io/netty/handler/ssl/util/FingerprintTrustManagerFactory;->checkTrusted",
	},
	"Squareup Certificate Pinning (OkHTTP<v3)": {
		"com/squareup/okhttp/CertificatePinner;->check",
	},
	"Squareup OkHostnameVerifier Pinning": {
		"com/squareup/okhttp/internal/tls/OkHostnameVerifier;->verify",
	},
	"Android WebViewClient SSL Pinning": {
		"android/webkit/WebViewClient;->onReceivedSslError",
	},
	"Apache Cordova WebViewClient Pinning": {
		"org/apache/cordova/CordovaWebViewClient;->onReceivedSslError",
	},
	"Boye AbstractVerifier Pinning": {
		"ch/boye/httpclientandroidlib/conn/ssl/AbstractVerifier;->verify",
	},
	"Apache AbstractVerifier Pinning": {
		"org/apache/http/conn/ssl/AbstractVerifier;->verify",
	},
	"Chromium Cronet Pinning": {
		"org/chromium/net/impl/CronetEngineBuilderImpl;->enablePublicKeyPinningBypassForLocalTrustAnchors",
		"org/chromium/net/impl/CronetEngineBuilderImpl;->addPublicKeyPins",
	},
	"Flutter HttpCertificatePinning": {
		"diefferson/http_certificate_pinning/HttpCertificatePinning;->checkConnexion",
	},
	"Flutter SslPinningPlugin": {
		"com/macif/plugin/sslpinningplugin/SslPinningPlugin;->checkConnexion",
	},
	"Custom Certificate Pinning": {
		"sha256/",
		"sha1/",
	},
}

// RootDetectionPatterns maps root detection libraries to smali patterns.
var RootDetectionPatterns = map[string][]string{
	"RootBeer": {
		"com/scottyab/rootbeer/RootBeer;",
		"com/scottyab/rootbeer/RootBeer;->isRooted",
		"com/scottyab/rootbeer/RootBeer;->detectRootManagementApps",
		"com/scottyab/rootbeer/RootBeer;->detectPotentiallyDangerousApps",
		"com/scottyab/rootbeer/RootBeer;->checkForSuBinary",
		"com/scottyab/rootbeer/RootBeer;->checkForBusyBoxBinary",
		"com/scottyab/rootbeer/RootBeer;->checkForDangerousProps",
		"com/scottyab/rootbeer/RootBeer;->checkForRWPaths",
		"com/scottyab/rootbeer/RootBeer;->detectTestKeys",
		"com/scottyab/rootbeer/RootBeerNative;",
	},
	"Jail Monkey (React Native)": {
		"com/gantman/react/RNJailMonkeyModule;",
		"com/gantman/react/RNJailMonkeyPackage;",
		"JailMonkeyModule",
		"isJailBroken",
		"canMockLocation",
		"isDebuggedMode",
	},
	"SafetyNet Attestation": {
		"com/google/android/gms/safetynet/SafetyNetApi;",
		"com/google/android/gms/safetynet/SafetyNetClient;",
		"com/google/android/gms/safetynet/SafetyNet;->getClient",
		"SafetyNetApi$AttestationResult",
	},
	"Play Integrity API": {
		"com/google/android/play/core/integrity/",
		"com/google/android/play/core/integrity/IntegrityManager;",
		"com/google/android/play/core/integrity/IntegrityTokenRequest;",
		"com/google/android/play/core/integrity/IntegrityTokenResponse;",
	},
	"Root Cloak": {
		"com/devadvance/rootcloak/",
		"com/devadvance/rootcloakplus/",
	},
	"Magisk Detection": {
		"com.topjohnwu.magisk",
		"/sbin/.magisk",
		"/sbin/magisk",
		"magisk",
		"MagiskHide",
	},
	"Generic Root Checks": {
		"/system/app/Superuser.apk",
		"/system/xbin/su",
		"/system/bin/su",
		"/sbin/su",
		"/data/local/xbin/su",
		"/data/local/bin/su",
		"/data/local/su",
		"com.noshufou.android.su",
		"com.thirdparty.superuser",
		"eu.chainfire.supersu",
		"com.koushikdutta.superuser",
		"com.zachspong.temprootremovejb",
		"com.ramdroid.appquarantine",
	},
	"BusyBox Detection": {
		"/system/xbin/busybox",
		"/system/bin/busybox",
		"busybox",
	},
	"System Property Checks": {
		"ro.debuggable",
		"ro.secure",
		"ro.build.selinux",
		"ro.build.tags",
		"test-keys",
	},
	"Frida Detection": {
		"frida-server",
		"libfrida-gadget",
		"frida-gadget",
		"/data/local/tmp/frida",
		"27042", // Frida default port
	},
	"Xposed Detection": {
		"de.robv.android.xposed",
		"de/robv/android/xposed/XposedBridge;",
		"XposedBridge",
		"XposedHelpers",
	},
}

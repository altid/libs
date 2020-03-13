#import "AltidLibPlugin.h"
#if __has_include(<altid_lib/altid_lib-Swift.h>)
#import <altid_lib/altid_lib-Swift.h>
#else
// Support project import fallback if the generated compatibility header
// is not copied when this plugin is created as a library.
// https://forums.swift.org/t/swift-static-libraries-dont-copy-generated-objective-c-header/19816
#import "altid_lib-Swift.h"
#endif

@implementation AltidLibPlugin
+ (void)registerWithRegistrar:(NSObject<FlutterPluginRegistrar>*)registrar {
  [SwiftAltidLibPlugin registerWithRegistrar:registrar];
}
@end

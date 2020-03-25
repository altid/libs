#import "AltidPlugin.h"
#if __has_include(<altid/altid-Swift.h>)
#import <altid/altid-Swift.h>
#else
// Support project import fallback if the generated compatibility header
// is not copied when this plugin is created as a library.
// https://forums.swift.org/t/swift-static-libraries-dont-copy-generated-objective-c-header/19816
#import "altid-Swift.h"
#endif

@implementation AltidPlugin
+ (void)registerWithRegistrar:(NSObject<FlutterPluginRegistrar>*)registrar {
  [SwiftAltidPlugin registerWithRegistrar:registrar];
}
@end

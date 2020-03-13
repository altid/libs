import 'dart:async';

import 'package:flutter/services.dart';

class AltidLib {
  static const MethodChannel _channel =
      const MethodChannel('altid_lib');

  static Future<String> get platformVersion async {
    final String version = await _channel.invokeMethod('getPlatformVersion');
    return version;
  }
}

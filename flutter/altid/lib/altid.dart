import 'dart:async';

import 'package:flutter/services.dart';

class Altid {
  static const MethodChannel _channel =
      const MethodChannel('altid');

  static Future<String> get platformVersion async {
    final String version = await _channel.invokeMethod('getPlatformVersion');
    return version;
  }
}

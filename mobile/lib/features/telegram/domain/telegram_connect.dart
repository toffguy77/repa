import 'package:freezed_annotation/freezed_annotation.dart';

part 'telegram_connect.freezed.dart';
part 'telegram_connect.g.dart';

@freezed
class TelegramConnectCode with _$TelegramConnectCode {
  const factory TelegramConnectCode({
    @JsonKey(name: 'connect_code') required String connectCode,
    required String instruction,
    @JsonKey(name: 'expires_at') required String expiresAt,
  }) = _TelegramConnectCode;

  factory TelegramConnectCode.fromJson(Map<String, dynamic> json) =>
      _$TelegramConnectCodeFromJson(json);
}

// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'telegram_connect.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_$TelegramConnectCodeImpl _$$TelegramConnectCodeImplFromJson(
        Map<String, dynamic> json) =>
    _$TelegramConnectCodeImpl(
      connectCode: json['connect_code'] as String,
      instruction: json['instruction'] as String,
      expiresAt: json['expires_at'] as String,
    );

Map<String, dynamic> _$$TelegramConnectCodeImplToJson(
        _$TelegramConnectCodeImpl instance) =>
    <String, dynamic>{
      'connect_code': instance.connectCode,
      'instruction': instance.instruction,
      'expires_at': instance.expiresAt,
    };

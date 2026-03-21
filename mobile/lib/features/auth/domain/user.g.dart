// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'user.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_$UserImpl _$$UserImplFromJson(Map<String, dynamic> json) => _$UserImpl(
      id: json['id'] as String,
      username: json['username'] as String,
      avatarUrl: json['avatar_url'] as String?,
      avatarEmoji: json['avatar_emoji'] as String?,
      birthYear: (json['birth_year'] as num?)?.toInt(),
      createdAt: json['created_at'] as String,
    );

Map<String, dynamic> _$$UserImplToJson(_$UserImpl instance) =>
    <String, dynamic>{
      'id': instance.id,
      'username': instance.username,
      'avatar_url': instance.avatarUrl,
      'avatar_emoji': instance.avatarEmoji,
      'birth_year': instance.birthYear,
      'created_at': instance.createdAt,
    };

import 'package:freezed_annotation/freezed_annotation.dart';

part 'user.freezed.dart';
part 'user.g.dart';

@freezed
class User with _$User {
  const factory User({
    required String id,
    required String username,
    @JsonKey(name: 'avatar_url') String? avatarUrl,
    @JsonKey(name: 'avatar_emoji') String? avatarEmoji,
    @JsonKey(name: 'birth_year') int? birthYear,
    @JsonKey(name: 'created_at') required String createdAt,
  }) = _User;

  factory User.fromJson(Map<String, dynamic> json) => _$UserFromJson(json);
}

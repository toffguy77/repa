import 'package:dio/dio.dart';
import '../../../core/api/api_client.dart';
import '../../../core/api/api_service.dart';
import '../domain/user.dart';

class AuthResult {
  final String token;
  final User user;

  AuthResult({required this.token, required this.user});
}

class AuthRepository {
  final ApiService _api;

  AuthRepository(this._api);

  Future<void> sendOtp(String phone) async {
    try {
      await _api.otpSend({'phone': phone});
    } on DioException catch (e) {
      throw parseError(e);
    }
  }

  Future<AuthResult> verifyOtp(String phone, String code) async {
    try {
      final response = await _api.otpVerify({'phone': phone, 'code': code});
      final data = response['data'] as Map<String, dynamic>;
      return AuthResult(
        token: data['token'] as String,
        user: User.fromJson(data['user'] as Map<String, dynamic>),
      );
    } on DioException catch (e) {
      throw parseError(e);
    }
  }

  Future<bool> checkUsername(String username) async {
    try {
      final response = await _api.checkUsername(username);
      final data = response['data'] as Map<String, dynamic>;
      return data['available'] as bool;
    } on DioException catch (e) {
      throw parseError(e);
    }
  }

  Future<User> updateProfile({
    String? username,
    String? avatarEmoji,
    int? birthYear,
  }) async {
    try {
      final body = <String, dynamic>{};
      if (username != null) body['username'] = username;
      if (avatarEmoji != null) body['avatar_emoji'] = avatarEmoji;
      if (birthYear != null) body['birth_year'] = birthYear;
      final response = await _api.updateProfile(body);
      final data = response['data'] as Map<String, dynamic>;
      return User.fromJson(data);
    } on DioException catch (e) {
      throw parseError(e);
    }
  }
}

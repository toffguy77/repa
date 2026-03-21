import 'package:dio/dio.dart';

class ApiService {
  final Dio _dio;

  ApiService(this._dio);

  // --- Auth ---

  Future<Map<String, dynamic>> otpSend(Map<String, String> body) async {
    final response = await _dio.post('/auth/otp/send', data: body);
    return response.data as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> otpVerify(Map<String, String> body) async {
    final response = await _dio.post('/auth/otp/verify', data: body);
    return response.data as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> appleAuth(Map<String, String> body) async {
    final response = await _dio.post('/auth/apple', data: body);
    return response.data as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> googleAuth(Map<String, String> body) async {
    final response = await _dio.post('/auth/google', data: body);
    return response.data as Map<String, dynamic>;
  }

  // --- Profile ---

  Future<Map<String, dynamic>> getMe() async {
    final response = await _dio.get('/auth/me');
    return response.data as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> checkUsername(String username) async {
    final response = await _dio.get('/auth/username/check',
        queryParameters: {'username': username});
    return response.data as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> updateProfile(
      Map<String, dynamic> body) async {
    final response = await _dio.put('/auth/profile', data: body);
    return response.data as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> uploadAvatar(MultipartFile file) async {
    final formData = FormData.fromMap({'file': file});
    final response = await _dio.post('/auth/avatar', data: formData);
    return response.data as Map<String, dynamic>;
  }

  // --- App ---

  Future<Map<String, dynamic>> appVersion() async {
    final response = await _dio.get('/auth/version');
    return response.data as Map<String, dynamic>;
  }
}

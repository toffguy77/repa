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

  // --- Groups ---

  Future<Map<String, dynamic>> createGroup(Map<String, dynamic> body) async {
    final response = await _dio.post('/groups', data: body);
    return response.data as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> listGroups() async {
    final response = await _dio.get('/groups');
    return response.data as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> getGroup(String id) async {
    final response = await _dio.get('/groups/$id');
    return response.data as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> joinPreview(String inviteCode) async {
    final response = await _dio.get('/groups/join/$inviteCode/preview');
    return response.data as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> joinGroup(String inviteCode) async {
    final response = await _dio.post('/groups/join/$inviteCode');
    return response.data as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> leaveGroup(String id) async {
    final response = await _dio.delete('/groups/$id/leave');
    return response.data as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> updateGroup(
      String id, Map<String, dynamic> body) async {
    final response = await _dio.patch('/groups/$id', data: body);
    return response.data as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> regenerateInviteLink(String id) async {
    final response = await _dio.post('/groups/$id/invite-link');
    return response.data as Map<String, dynamic>;
  }

  // --- Voting ---

  Future<Map<String, dynamic>> getVotingSession(String seasonId) async {
    final response = await _dio.get('/seasons/$seasonId/voting-session');
    return response.data as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> castVote(
      String seasonId, Map<String, String> body) async {
    final response = await _dio.post('/seasons/$seasonId/votes', data: body);
    return response.data as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> getVotingProgress(String seasonId) async {
    final response = await _dio.get('/seasons/$seasonId/progress');
    return response.data as Map<String, dynamic>;
  }

  // --- Reveal ---

  Future<Map<String, dynamic>> getReveal(String seasonId) async {
    final response = await _dio.get('/seasons/$seasonId/reveal');
    return response.data as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> getMembersCards(String seasonId) async {
    final response = await _dio.get('/seasons/$seasonId/members-cards');
    return response.data as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> openHidden(String seasonId) async {
    final response =
        await _dio.post('/seasons/$seasonId/reveal/open-hidden');
    return response.data as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> getMyCardUrl(String seasonId) async {
    final response = await _dio.get('/seasons/$seasonId/my-card-url');
    return response.data as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> getDetector(String seasonId) async {
    final response = await _dio.get('/seasons/$seasonId/detector');
    return response.data as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> buyDetector(String seasonId) async {
    final response = await _dio.post('/seasons/$seasonId/detector');
    return response.data as Map<String, dynamic>;
  }
}

import '../../../core/api/api_service.dart';
import '../domain/profile.dart';

class ProfileRepository {
  final ApiService _api;

  ProfileRepository(this._api);

  Future<MemberProfile> getMemberProfile(String groupId, String userId) async {
    final response = await _api.getMemberProfile(groupId, userId);
    final data = response['data'] as Map<String, dynamic>;
    return MemberProfile.fromJson(data);
  }
}

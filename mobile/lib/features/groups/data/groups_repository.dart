import 'package:dio/dio.dart';
import '../../../core/api/api_client.dart';
import '../../../core/api/api_service.dart';
import '../domain/group.dart';

class CreateGroupResult {
  final Group group;
  final String inviteUrl;

  CreateGroupResult({required this.group, required this.inviteUrl});
}

class GroupsRepository {
  final ApiService _api;

  GroupsRepository(this._api);

  /// [kindOnly] is sent only when the creator actually chose a value. Omitting it lets the server
  /// apply its age-based default, which is the only place that rule lives — a client that always sent
  /// its own switch value would override it, and would get it wrong whenever it does not know the
  /// creator's birth year.
  Future<CreateGroupResult> createGroup({
    required String name,
    required List<String> categories,
    String? telegramUsername,
    bool? kindOnly,
  }) async {
    try {
      final body = <String, dynamic>{
        'name': name,
        'categories': categories,
      };
      if (kindOnly != null) {
        body['kind_only'] = kindOnly;
      }
      if (telegramUsername != null && telegramUsername.isNotEmpty) {
        body['telegram_username'] = telegramUsername;
      }
      final response = await _api.createGroup(body);
      final data = response['data'] as Map<String, dynamic>;
      return CreateGroupResult(
        group: Group.fromJson(data['group'] as Map<String, dynamic>),
        inviteUrl: data['invite_url'] as String,
      );
    } on DioException catch (e) {
      throw parseError(e);
    }
  }

  /// Changes the group's kind-only setting. Admin only, enforced by the server.
  ///
  /// Applies from the next season: the open one is left as members found it.
  Future<Group> setKindOnly(String groupId, bool kindOnly) async {
    try {
      final response = await _api.updateGroup(groupId, {'kind_only': kindOnly});
      final data = response['data'] as Map<String, dynamic>;
      return Group.fromJson(data['group'] as Map<String, dynamic>);
    } on DioException catch (e) {
      throw parseError(e);
    }
  }

  Future<List<GroupListItem>> listGroups() async {
    try {
      final response = await _api.listGroups();
      final data = response['data'] as Map<String, dynamic>;
      final list = data['groups'] as List<dynamic>;
      return list
          .map((e) => GroupListItem.fromJson(e as Map<String, dynamic>))
          .toList();
    } on DioException catch (e) {
      throw parseError(e);
    }
  }

  Future<GroupDetail> getGroup(String id) async {
    try {
      final response = await _api.getGroup(id);
      final data = response['data'] as Map<String, dynamic>;
      return GroupDetail.fromJson(data);
    } on DioException catch (e) {
      throw parseError(e);
    }
  }

  Future<JoinPreview> joinPreview(String inviteCode) async {
    try {
      final response = await _api.joinPreview(inviteCode);
      final data = response['data'] as Map<String, dynamic>;
      return JoinPreview.fromJson(data);
    } on DioException catch (e) {
      throw parseError(e);
    }
  }

  Future<Group> joinGroup(
    String inviteCode, {
    String? source,
    String? referrerId,
  }) async {
    try {
      final response = await _api.joinGroup(
        inviteCode,
        source: source,
        referrerId: referrerId,
      );
      final data = response['data'] as Map<String, dynamic>;
      return Group.fromJson(data['group'] as Map<String, dynamic>);
    } on DioException catch (e) {
      throw parseError(e);
    }
  }

  Future<void> blockMember(String userId) async {
    try {
      await _api.blockMember(userId);
    } on DioException catch (e) {
      throw parseError(e);
    }
  }

  Future<void> unblockMember(String userId) async {
    try {
      await _api.unblockMember(userId);
    } on DioException catch (e) {
      throw parseError(e);
    }
  }

  Future<void> reportMember(
    String userId, {
    required String groupId,
    String? reason,
  }) async {
    try {
      await _api.reportMember(userId, groupId: groupId, reason: reason);
    } on DioException catch (e) {
      throw parseError(e);
    }
  }

  Future<void> removeMember(String groupId, String userId) async {
    try {
      await _api.removeMember(groupId, userId);
    } on DioException catch (e) {
      throw parseError(e);
    }
  }

  Future<void> leaveGroup(String id, {bool permanent = false}) async {
    try {
      await _api.leaveGroup(id, permanent: permanent);
    } on DioException catch (e) {
      throw parseError(e);
    }
  }

  Future<String> regenerateInviteLink(String id) async {
    try {
      final response = await _api.regenerateInviteLink(id);
      final data = response['data'] as Map<String, dynamic>;
      return data['invite_url'] as String;
    } on DioException catch (e) {
      throw parseError(e);
    }
  }
}

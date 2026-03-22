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

  Future<CreateGroupResult> createGroup({
    required String name,
    required List<String> categories,
    String? telegramUsername,
  }) async {
    try {
      final body = <String, dynamic>{
        'name': name,
        'categories': categories,
      };
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

  Future<Group> joinGroup(String inviteCode) async {
    try {
      final response = await _api.joinGroup(inviteCode);
      final data = response['data'] as Map<String, dynamic>;
      return Group.fromJson(data['group'] as Map<String, dynamic>);
    } on DioException catch (e) {
      throw parseError(e);
    }
  }

  Future<void> leaveGroup(String id) async {
    try {
      await _api.leaveGroup(id);
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

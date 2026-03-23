import 'package:dio/dio.dart';
import '../../../core/api/api_client.dart';
import '../../../core/api/api_service.dart';
import '../domain/telegram_connect.dart';

class TelegramRepository {
  final ApiService _api;

  TelegramRepository(this._api);

  Future<TelegramConnectCode> generateCode(String groupId) async {
    try {
      final response = await _api.generateTelegramCode(groupId);
      final data = response['data'] as Map<String, dynamic>;
      return TelegramConnectCode.fromJson(data);
    } on DioException catch (e) {
      throw parseError(e);
    }
  }

  Future<void> disconnect(String groupId) async {
    try {
      await _api.disconnectTelegram(groupId);
    } on DioException catch (e) {
      throw parseError(e);
    }
  }

  Future<void> shareToTelegram(String seasonId) async {
    try {
      await _api.shareToTelegram(seasonId);
    } on DioException catch (e) {
      throw parseError(e);
    }
  }
}

import 'package:dio/dio.dart';

import '../../../core/api/api_client.dart';
import '../../../core/api/api_service.dart';
import '../domain/chronicle.dart';

class ChronicleRepository {
  final ApiService _api;

  ChronicleRepository(this._api);

  Future<Chronicle> getChronicle(String groupId) async {
    try {
      final response = await _api.getGroupChronicle(groupId);
      return Chronicle.fromJson(response['data'] as Map<String, dynamic>);
    } on DioException catch (e) {
      throw parseError(e);
    }
  }
}

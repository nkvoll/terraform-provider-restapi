package restapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceRestAPI() *schema.Resource {
	// Consider data sensitive if env variables is set to true.
	isDataSensitive, _ := strconv.ParseBool(GetEnvOrDefault("API_DATA_IS_SENSITIVE", "false"))

	return &schema.Resource{
		CreateContext: func(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
			return diag.FromErr(resourceRestAPICreate(ctx, data, i))
		},
		ReadContext: func(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
			return diag.FromErr(resourceRestAPIRead(ctx, data, i))
		},
		UpdateContext: func(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
			return diag.FromErr(resourceRestAPIUpdate(ctx, data, i))
		},
		DeleteContext: func(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
			return diag.FromErr(resourceRestAPIDelete(ctx, data, i))
		},

		Description: "Acting as a wrapper of cURL, this object supports POST, GET, PUT and DELETE on the specified url",

		Importer: &schema.ResourceImporter{
			StateContext: resourceRestAPIImport,
		},

		Schema: map[string]*schema.Schema{
			"path": {
				Type:        schema.TypeString,
				Description: "The API path on top of the base URL set in the provider that represents objects of this type on the API server.",
				Required:    true,
			},
			"create_path": {
				Type:        schema.TypeString,
				Description: "Defaults to `path`. The API path that represents where to CREATE (POST) objects of this type on the API server. The string `{id}` will be replaced with the terraform ID of the object if the data contains the `id_attribute`.",
				Optional:    true,
			},
			"read_path": {
				Type:        schema.TypeString,
				Description: "Defaults to `path/{id}`. The API path that represents where to READ (GET) objects of this type on the API server. The string `{id}` will be replaced with the terraform ID of the object.",
				Optional:    true,
			},
			"update_path": {
				Type:        schema.TypeString,
				Description: "Defaults to `path/{id}`. The API path that represents where to UPDATE (PUT) objects of this type on the API server. The string `{id}` will be replaced with the terraform ID of the object.",
				Optional:    true,
			},
			"create_method": {
				Type:        schema.TypeString,
				Description: "Defaults to `create_method` set on the provider. Allows per-resource override of `create_method` (see `create_method` provider config documentation)",
				Optional:    true,
			},
			"read_method": {
				Type:        schema.TypeString,
				Description: "Defaults to `read_method` set on the provider. Allows per-resource override of `read_method` (see `read_method` provider config documentation)",
				Optional:    true,
			},
			"update_method": {
				Type:        schema.TypeString,
				Description: "Defaults to `update_method` set on the provider. Allows per-resource override of `update_method` (see `update_method` provider config documentation)",
				Optional:    true,
			},
			"destroy_method": {
				Type:        schema.TypeString,
				Description: "Defaults to `destroy_method` set on the provider. Allows per-resource override of `destroy_method` (see `destroy_method` provider config documentation)",
				Optional:    true,
			},
			"destroy_path": {
				Type:        schema.TypeString,
				Description: "Defaults to `path/{id}`. The API path that represents where to DESTROY (DELETE) objects of this type on the API server. The string `{id}` will be replaced with the terraform ID of the object.",
				Optional:    true,
			},
			"id_attribute": {
				Type:        schema.TypeString,
				Description: "Defaults to `id_attribute` set on the provider. Allows per-resource override of `id_attribute` (see `id_attribute` provider config documentation)",
				Optional:    true,
			},
			"object_id": {
				Type:        schema.TypeString,
				Description: "Defaults to the id learned by the provider during normal operations and `id_attribute`. Allows you to set the id manually. This is used in conjunction with the `*_path` attributes.",
				Optional:    true,
			},
			"data": {
				Type:        schema.TypeString,
				Description: "Valid JSON object that this provider will manage with the API server.",
				Required:    true,
				Sensitive:   isDataSensitive,
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(string)
					if v != "" {
						data := make(map[string]interface{})
						err := json.Unmarshal([]byte(v), &data)
						if err != nil {
							errs = append(errs, fmt.Errorf("data attribute is invalid JSON: %v", err))
						}
					}
					return warns, errs
				},
			},
			"debug": {
				Type:        schema.TypeBool,
				Description: "Whether to emit verbose debug output while working with the API object on the server.",
				Optional:    true,
			},
			"read_search": {
				Type:        schema.TypeMap,
				Description: "Custom search for `read_path`. This map will take `search_key`, `search_value`, `results_key` and `query_string` (see datasource config documentation)",
				Optional:    true,
			},
			"query_string": {
				Type:        schema.TypeString,
				Description: "Query string to be included in the path",
				Optional:    true,
			},
			"read_query_string": {
				Type:        schema.TypeString,
				Description: "Query string to be included in the path",
				Optional:    true,
			},
			"create_query_string": {
				Type:        schema.TypeString,
				Description: "Query string to be included in the path",
				Optional:    true,
			},
			"update_query_string": {
				Type:        schema.TypeString,
				Description: "Query string to be included in the path",
				Optional:    true,
			},
			"delete_query_string": {
				Type:        schema.TypeString,
				Description: "Query string to be included in the path",
				Optional:    true,
			},
			"force_new": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				ForceNew:    true,
				Description: "Any changes to these values will result in recreating the resource instead of updating.",
			},
			"update_data": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Valid JSON object to pass during to update requests.",
				Sensitive:   isDataSensitive,
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(string)
					if v != "" {
						data := make(map[string]interface{})
						err := json.Unmarshal([]byte(v), &data)
						if err != nil {
							errs = append(errs, fmt.Errorf("update_data attribute is invalid JSON: %v", err))
						}
					}
					return warns, errs
				},
			},
			"destroy_data": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Valid JSON object to pass during to destroy requests.",
				Sensitive:   isDataSensitive,
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(string)
					if v != "" {
						data := make(map[string]interface{})
						err := json.Unmarshal([]byte(v), &data)
						if err != nil {
							errs = append(errs, fmt.Errorf("destroy_data attribute is invalid JSON: %v", err))
						}
					}
					return warns, errs
				},
			},
			"ignore_changes_to": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "A list of fields to which remote changes will be ignored. For example, an API might add or remove metadata, such as a 'last_modified' field, which Terraform should not attempt to correct. To ignore changes to nested fields, use the dot syntax: 'metadata.timestamp'",
				Sensitive:   isDataSensitive,
				// TODO ValidateFunc not supported for lists, but should probably validate that the ignore paths are valid
			},
			"ignore_all_server_changes": {
				Type:        schema.TypeBool,
				Description: "By default Terraform will attempt to revert changes to remote resources. Set this to 'true' to ignore any remote changes. Default: false",
				Optional:    true,
				Default:     false,
			},
			"drift_fields": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "An object that matches the structure of the data to which remote changes will be considered when detecting drift. Default to the empty object which means all changes are included. ",
				Sensitive:   isDataSensitive,
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(string)
					if v != "" {
						data := make(map[string]interface{})
						err := json.Unmarshal([]byte(v), &data)
						if err != nil {
							errs = append(errs, fmt.Errorf("destroy_data attribute is invalid JSON: %v", err))
						}
					}
					return warns, errs
				},
			},
			"drift_fields_from_data": {
				Type:        schema.TypeBool,
				Description: "Set this to 'true' to use the data as drift fields to make only explicitly set fields are checked for drift. Default: false",
				Optional:    true,
				Default:     false,
			},
		}, /* End schema */

	}
}

/*
Since there is nothing in the ResourceData structure other

	than the "id" passed on the command line, we have to use an opinionated
	view of the API paths to figure out how to read that object
	from the API
*/
func resourceRestAPIImport(ctx context.Context, d *schema.ResourceData, meta interface{}) (imported []*schema.ResourceData, err error) {
	input := d.Id()

	hasTrailingSlash := strings.HasSuffix(input, "/")
	var n int
	if hasTrailingSlash {
		n = strings.LastIndex(input[0:len(input)-1], "/")
	} else {
		n = strings.LastIndex(input, "/")
	}

	if n == -1 {
		return imported, fmt.Errorf("invalid path to import api_object '%s' - must be /<full path from server root>/<object id>", input)
	}

	path := input[0:n]
	d.Set("path", path)

	var id string
	if hasTrailingSlash {
		id = input[n+1 : len(input)-1]
	} else {
		id = input[n+1:]
	}

	d.Set("data", fmt.Sprintf(`{ "id": "%s" }`, id))
	d.SetId(id)

	/* Troubleshooting is hard enough. Emit log messages so TF_LOG
	   has useful information in case an import isn't working */
	d.Set("debug", true)

	obj, err := makeAPIObject(d, meta)
	if err != nil {
		return imported, err
	}
	log.Printf("resource_api_object.go: Import routine called. Object built:\n%s\n", obj.toString())

	err = obj.readObject(ctx)
	if err == nil {
		//setResourceState(obj, d)
		/* Data that we set in the state above must be passed along
		   as an item in the stack of imported data */
		imported = append(imported, d)
	}

	return imported, err
}

func resourceRestAPICreate(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	obj, err := makeAPIObject(d, meta)
	if err != nil {
		return err
	}
	log.Printf("resource_api_object.go: Create routine called. Object built:\n%s\n", obj.toString())

	err = obj.createObject(ctx)
	if err == nil {
		/* Setting terraform ID tells terraform the object was created or it exists */
		d.SetId(obj.id)
		//setResourceState(obj, d)
		/* Only set during create for APIs that don't return sensitive data on subsequent retrieval */
		//d.Set("create_response", obj.apiResponse)
	}
	return err
}

func resourceRestAPIRead(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	obj, err := makeAPIObject(d, meta)
	if err != nil {
		if strings.Contains(err.Error(), "error parsing data provided") {
			log.Printf("resource_api_object.go: WARNING! The data passed from Terraform's state is invalid! %v", err)
			log.Printf("resource_api_object.go: Continuing with partially constructed object...")
		} else {
			return err
		}
	}
	log.Printf("resource_api_object.go: Read routine called. Object built:\n%s\n", obj.toString())

	err = obj.readObject(ctx)
	if err == nil {
		/* Setting terraform ID tells terraform the object was created or it exists */
		log.Printf("resource_api_object.go: Read resource. Returned id is '%s'\n", obj.id)
		d.SetId(obj.id)

		//setResourceState(obj, d)

		// Check whether the remote resource has changed.
		if !(d.Get("ignore_all_server_changes")).(bool) {
			ignoreList := []string{}
			v, ok := d.GetOk("ignore_changes_to")
			if ok {
				for _, s := range v.([]interface{}) {
					ignoreList = append(ignoreList, s.(string))
				}
			}

			var driftFields map[string]interface{}
			if v, ok := d.GetOk("drift_fields_from_data"); ok {
				if v.(bool) {
					driftFields = obj.data
				}
			}

			if v, ok = d.GetOk("drift_fields"); ok {
				if err := json.Unmarshal([]byte(v.(string)), &driftFields); err != nil {
					return err
				}
			}

			// This checks if there were any changes to the remote resource that will need to be corrected
			// by comparing the current state with the response returned by the api.
			modifiedResource, hasDifferences := getDelta(obj.data, obj.apiData, ignoreList, driftFields)

			if hasDifferences {
				log.Printf("resource_api_object.go: Found differences in remote resource\n")
				encoded, err := json.Marshal(modifiedResource)
				if err != nil {
					return err
				}
				jsonString := string(encoded)
				if err := d.Set("data", jsonString); err != nil {
					return err
				}
			}
		}

	}
	return err
}

func resourceRestAPIUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	obj, err := makeAPIObject(d, meta)
	if err != nil {
		return err
	}

	/* If copy_keys is not empty, we have to grab the latest
	   data so we can copy anything needed before the update */
	client := meta.(*APIClient)
	if len(client.copyKeys) > 0 {
		err = obj.readObject(ctx)
		if err != nil {
			return err
		}
	}

	log.Printf("resource_api_object.go: Update routine called. Object built:\n%s\n", obj.toString())

	err = obj.updateObject(ctx)
	if err == nil {
		//setResourceState(obj, d)
	}
	return err
}

func resourceRestAPIDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	obj, err := makeAPIObject(d, meta)
	if err != nil {
		return err
	}
	log.Printf("resource_api_object.go: Delete routine called. Object built:\n%s\n", obj.toString())

	err = obj.deleteObject(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			/* 404 means it doesn't exist. Call that good enough */
			err = nil
		}
	}
	return err
}

/*
Simple helper routine to build an api_object struct

	for the various calls terraform will use. Unfortunately,
	terraform cannot just reuse objects, so each CRUD operation
	results in a new object created
*/
func makeAPIObject(d *schema.ResourceData, meta interface{}) (*APIObject, error) {
	opts, err := buildAPIObjectOpts(d)
	if err != nil {
		return nil, err
	}

	caller := "unknown"
	pc, _, _, ok := runtime.Caller(1)
	details := runtime.FuncForPC(pc)
	if ok && details != nil {
		parts := strings.Split(details.Name(), ".")
		caller = parts[len(parts)-1]
	}
	log.Printf("resource_rest_api.go: Constructing new APIObject in makeAPIObject (called by %s)", caller)

	obj, err := NewAPIObject(meta.(*APIClient), opts)

	return obj, err
}

func buildAPIObjectOpts(d *schema.ResourceData) (*apiObjectOpts, error) {
	opts := &apiObjectOpts{
		path: d.Get("path").(string),
	}

	/* Allow user to override provider-level id_attribute */
	if v, ok := d.GetOk("id_attribute"); ok {
		opts.idAttribute = v.(string)
	}

	/* Allow user to specify the ID manually */
	if v, ok := d.GetOk("object_id"); ok {
		opts.id = v.(string)
	} else {
		/* If not specified, see if terraform has an ID */
		opts.id = d.Id()
	}

	log.Printf("resource_rest_api.go: buildAPIObjectOpts routine called for id '%s'\n", opts.id)

	if v, ok := d.GetOk("create_path"); ok {
		opts.postPath = v.(string)
	}
	if v, ok := d.GetOk("read_path"); ok {
		opts.getPath = v.(string)
	}
	if v, ok := d.GetOk("update_path"); ok {
		opts.putPath = v.(string)
	}
	if v, ok := d.GetOk("create_method"); ok {
		opts.createMethod = v.(string)
	}
	if v, ok := d.GetOk("read_method"); ok {
		opts.readMethod = v.(string)
	}
	if v, ok := d.GetOk("update_method"); ok {
		opts.updateMethod = v.(string)
	}
	if v, ok := d.GetOk("update_data"); ok {
		opts.updateData = v.(string)
	}
	if v, ok := d.GetOk("destroy_method"); ok {
		opts.destroyMethod = v.(string)
	}
	if v, ok := d.GetOk("destroy_data"); ok {
		opts.destroyData = v.(string)
	}
	if v, ok := d.GetOk("destroy_path"); ok {
		opts.deletePath = v.(string)
	}
	if v, ok := d.GetOk("query_string"); ok {
		opts.queryString = v.(string)
	}
	if v, ok := d.GetOk("read_query_string"); ok {
		opts.readQueryString = v.(string)
	}
	if v, ok := d.GetOk("create_query_string"); ok {
		opts.createQueryString = v.(string)
	}
	if v, ok := d.GetOk("update_query_string"); ok {
		opts.updateQueryString = v.(string)
	}
	if v, ok := d.GetOk("destroy_query_string"); ok {
		opts.destroyQueryString = v.(string)
	}

	readSearch := expandReadSearch(d.Get("read_search").(map[string]interface{}))
	opts.readSearch = readSearch

	opts.data = d.Get("data").(string)
	opts.debug = d.Get("debug").(bool)

	return opts, nil
}

func expandReadSearch(v map[string]interface{}) (readSearch map[string]string) {
	readSearch = make(map[string]string)
	for key, val := range v {
		readSearch[key] = val.(string)
	}

	return
}
